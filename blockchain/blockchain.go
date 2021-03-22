package blockchain

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"math/big"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/blockchain/forging"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/stake"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/store"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Blockchain struct {
	ChainData             unsafe.Pointer       //using atomic storePointer
	Sync                  bool                 `json:"-"`
	UpdateChannel         chan uint64          `json:"-"`
	UpdateNewChainChannel chan *BlockchainData `json:"-"`
	forging               *forging.Forging     `json:"-"`
	mempool               *mempool.Mempool     `json:"-"`
	mutex                 sync.Mutex           `json:"-"` //writing mutex
	sync.RWMutex          `json:"-"`
}

func (chain *Blockchain) AddBlocks(blocksComplete []*block_complete.BlockComplete, calledByForging bool) (result bool, err error) {

	result = false
	if len(blocksComplete) == 0 {
		err = errors.New("Blocks length is ZERO")
		return
	}

	for _, blkComplete := range blocksComplete {
		blkComplete.Validate()
		blkComplete.Verify()
	}

	//avoid processing the same function twice
	chain.mutex.Lock()

	chainData := (*BlockchainData)(atomic.LoadPointer(&chain.ChainData))
	gui.Info(fmt.Sprintf("Including blocks %d ... %d", chainData.Height, chainData.Height+uint64(len(blocksComplete))))

	//chain.RLock() is not required because it is guaranteed that no other thread is writing now in the chain
	var newChainData = &BlockchainData{
		Hash:               chainData.Hash,
		PrevHash:           chainData.PrevHash,
		KernelHash:         chainData.KernelHash,
		PrevKernelHash:     chainData.PrevKernelHash,
		Height:             chainData.Height,
		Timestamp:          chainData.Timestamp,
		Target:             chainData.Target,
		BigTotalDifficulty: chainData.BigTotalDifficulty,
		Transactions:       chainData.Transactions,
	}
	mainChainBigTotalDifficulty := chainData.BigTotalDifficulty

	var accs *accounts.Accounts
	var toks *tokens.Tokens

	boltTx, err := store.StoreBlockchain.DB.Begin(true)
	if err != nil {
		return
	}

	insertedBlocks := []*block_complete.BlockComplete{}
	insertedTxHashes := [][]byte{}

	var writer *bolt.Bucket
	savedBlock := false
	//remove blocks which are different
	removedTxHashes := make(map[string][]byte)
	removedTx := [][]byte{}
	removedBlocksHeights := []uint64{}

	func() {

		defer func() {

			err = helpers.ConvertRecoverError(recover())

			//recover, but in case the chain was correctly saved and the mewChainDifficulty is higher than
			//we should store it
			if savedBlock && mainChainBigTotalDifficulty.Cmp(newChainData.BigTotalDifficulty) < 0 {

				err = nil
				newChainData.saveBlockchain(writer)

				for _, removedBlock := range removedBlocksHeights {
					chain.deleteUnusedBlocksComplete(writer, removedBlock, accs, toks)
				}
				for txHash := range removedTxHashes {
					data := writer.Get([]byte("tx" + txHash))
					removedTx = append(removedTx, data)
					writer.Delete([]byte("tx" + txHash))
				}

				accs.Rollback()
				toks.Rollback()
				accs.WriteToStore()
				toks.WriteToStore()

				chain.Lock()
				err = boltTx.Commit()
				if err == nil {
					atomic.StorePointer(&chain.ChainData, unsafe.Pointer(newChainData))
				}
				chain.Unlock()

			} else {

				if err2 := boltTx.Rollback(); err2 != nil {
					gui.Error("Error rollback chain")
					err = errors.New("Error rollback chain")
				}

			}

		}()

		writer = boltTx.Bucket([]byte("Chain"))

		accs = accounts.NewAccounts(boltTx)
		toks = tokens.NewTokens(boltTx)

		//let's filter existing blocks
		for i := len(blocksComplete) - 1; i >= 0; i-- {

			blkComplete := blocksComplete[i]

			if blkComplete.Block.Height < newChainData.Height {
				hash := chain.LoadBlockHash(writer, blkComplete.Block.Height)
				if bytes.Equal(hash, blkComplete.Block.Bloom.Hash) {
					blocksComplete = blocksComplete[i+1:]
					break
				}
			}

		}

		if len(blocksComplete) == 0 {
			panic("blocks are identical now")
		}

		firstBlockComplete := blocksComplete[0]
		if firstBlockComplete.Block.Height < newChainData.Height {
			for i := newChainData.Height - 1; i >= newChainData.Height; i-- {
				removedBlocksHeights = append(removedBlocksHeights, 0)
				copy(removedBlocksHeights[1:], removedBlocksHeights)
				removedBlocksHeights[0] = i

				chain.removeBlockComplete(writer, i, removedTxHashes, accs, toks)
			}
		}

		if blocksComplete[0].Block.Height != newChainData.Height {
			panic("First Block has is not matching")
		}

		if !bytes.Equal(firstBlockComplete.Block.PrevHash, newChainData.Hash) {
			panic("First block hash is not matching chain hash")
		}

		if !bytes.Equal(firstBlockComplete.Block.PrevKernelHash, newChainData.KernelHash) {
			panic("First block kernel hash is not matching chain prev kerneh lash")
		}

		for i, blkComplete := range blocksComplete {

			//check block height
			if blkComplete.Block.Height != newChainData.Height {
				panic("Block Height is not right!")
			}

			//check blkComplete balance
			var stakingAmount uint64
			if blkComplete.Block.Height > 0 {

				acc := accs.GetAccount(blkComplete.Block.Forger)
				if acc == nil || !acc.HasDelegatedStake() {
					panic("Forger Account deson't exist or hasn't delegated stake")
				}
				stakingAmount = acc.GetDelegatedStakeAvailable(blkComplete.Block.Height)

				if !bytes.Equal(blkComplete.Block.DelegatedPublicKeyHash, acc.DelegatedStake.DelegatedPublicKeyHash) {
					panic("Block Staking Delegated Public Key is not matching")
				}

			}

			if blkComplete.Block.StakingAmount > stakingAmount {
				panic("Block Staking Amount doesn't match")
			}

			if blkComplete.Block.StakingAmount < stake.GetRequiredStake(blkComplete.Block.Height) {
				panic("Delegated stake ready amount is not enought")
			}

			if difficulty.CheckKernelHashBig(blkComplete.Block.Bloom.KernelHash, newChainData.Target) != true {
				panic("KernelHash Difficulty is not met")
			}

			//already verified for i == 0
			if i > 0 {
				if !bytes.Equal(blkComplete.Block.PrevHash, newChainData.Hash) {
					panic("PrevHash doesn't match Genesis prevHash")
				}
				if !bytes.Equal(blkComplete.Block.PrevKernelHash, newChainData.KernelHash) {
					panic("PrevHash doesn't match Genesis prevKernelHash")
				}
			}

			if blkComplete.Block.Timestamp < newChainData.Timestamp {
				panic("Timestamp has to be greather than the last timestmap")
			}

			if blkComplete.Block.Timestamp > uint64(time.Now().UTC().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
				panic("Timestamp is too much into the future")
			}

			blkComplete.IncludeBlockComplete(accs, toks)

			//to detect if the savedBlock was done correctly
			savedBlock = false

			newTransactionsSaved := chain.saveBlockComplete(writer, blkComplete, blkComplete.Block.Bloom.Hash, removedTxHashes, accs, toks)

			if len(removedBlocksHeights) > 0 {
				removedBlocksHeights = removedBlocksHeights[1:]
			}

			accs.Commit() //it will commit the changes but not save them
			toks.Commit() //it will commit the changes but not save them

			newChainData.PrevHash = newChainData.Hash
			newChainData.Hash = blkComplete.Block.Bloom.Hash
			newChainData.PrevKernelHash = newChainData.KernelHash
			newChainData.KernelHash = blkComplete.Block.Bloom.KernelHash
			newChainData.Timestamp = blkComplete.Block.Timestamp

			difficultyBigInt := difficulty.ConvertTargetToDifficulty(newChainData.Target)
			newChainData.BigTotalDifficulty = new(big.Int).Add(newChainData.BigTotalDifficulty, difficultyBigInt)
			newChainData.saveTotalDifficultyExtra(writer)

			newChainData.Target = newChainData.computeNextTargetBig(writer)

			newChainData.Height += 1
			newChainData.Transactions += uint64(len(blkComplete.Txs))
			insertedBlocks = append(insertedBlocks, blkComplete)

			for _, txHashId := range newTransactionsSaved {
				insertedTxHashes = append(insertedTxHashes, txHashId)
			}

			writer.Put([]byte("chainHash"), newChainData.Hash)
			writer.Put([]byte("chainPrevHash"), newChainData.PrevHash)
			writer.Put([]byte("chainKernelHash"), newChainData.KernelHash)
			writer.Put([]byte("chainPrevKernelHash"), newChainData.PrevKernelHash)

			buf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutUvarint(buf, newChainData.Height)
			writer.Put([]byte("chainHeight"), buf[:n])

			savedBlock = true
		}

	}()

	chain.mutex.Unlock()

	if err != nil {
		if calledByForging {
			chain.createNextBlockForForging()
		}
		return
	}

	gui.Warning("-------------------------------------------")
	gui.Warning(fmt.Sprintf("Including blocks %d | TXs: %d | Hash %s", len(insertedBlocks), len(insertedTxHashes), hex.EncodeToString(chainData.Hash)))
	gui.Warning("-------------------------------------------")
	newChainData.updateChainInfo()

	chain.UpdateChannel <- newChainData.Height //sending 1
	chain.UpdateNewChainChannel <- newChainData

	//accs will only be read only
	chain.forging.Wallet.UpdateBalanceChanges(accs)

	//create next block and the workers will be automatically reset
	chain.createNextBlockForForging()

	for _, txData := range removedTx {
		tx := transaction.Transaction{}
		tx.Deserialize(helpers.NewBufferReader(txData), true)
		chain.mempool.AddTxToMemPoolSilent(&tx, newChainData.Height, false)
	}

	for _, txHash := range insertedTxHashes {
		chain.mempool.Delete(txHash)
	}

	chain.mempool.UpdateWork(newChainData.Hash, newChainData.Height)

	result = true
	return

}

func BlockchainInit(forging *forging.Forging, mempool *mempool.Mempool) (chain *Blockchain) {

	gui.Log("Blockchain init...")

	genesis.GenesisInit()

	chain = &Blockchain{
		forging:               forging,
		mempool:               mempool,
		Sync:                  false,
		UpdateChannel:         make(chan uint64),
		UpdateNewChainChannel: make(chan *BlockchainData),
	}

	success, err := chain.loadBlockchain()
	if err != nil {
		panic(err)
	}

	if !success {
		chain.init()
	}

	chainData := (*BlockchainData)(atomic.LoadPointer(&chain.ChainData))
	chainData.updateChainInfo()

	chain.initForging()

	return
}

func (chain *Blockchain) Close() {
}
