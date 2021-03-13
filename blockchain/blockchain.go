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
	"pandora-pay/blockchain/block"
	"pandora-pay/blockchain/block/difficulty"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/stake"
	"pandora-pay/forging"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/store"
	"strconv"
	"sync"
	"time"
)

type Blockchain struct {
	Hash           []byte //32
	PrevHash       []byte //32
	KernelHash     []byte //32
	PrevKernelHash []byte //32
	Height         uint64
	Timestamp      uint64

	Target             *big.Int
	BigTotalDifficulty *big.Int

	Transactions uint64 //count of the number of txs

	Sync bool `json:"-"`

	UpdateChannel         chan uint64      `json:"-"`
	UpdateNewChainChannel chan *Blockchain `json:"-"`

	forging *forging.Forging `json:"-"`
	mempool *mempool.MemPool `json:"-"`

	mutex        sync.Mutex `json:"-"`
	sync.RWMutex `json:"-"`
}

func (chain *Blockchain) AddBlocks(blocksComplete []*block.BlockComplete, calledByForging bool) (result bool, err error) {

	result = false
	if len(blocksComplete) == 0 {
		err = errors.New("Blocks length is ZERO")
		return
	}

	//avoid processing the same function twice
	chain.mutex.Lock()

	gui.Info(fmt.Sprintf("Including blocks %d ... %d", chain.Height, chain.Height+uint64(len(blocksComplete))))

	//chain.RLock() is not required because it is guaranteed that no other thread is writing now in the chain
	var newChain = Blockchain{
		Hash:               chain.Hash,
		PrevHash:           chain.PrevHash,
		KernelHash:         chain.KernelHash,
		PrevKernelHash:     chain.PrevKernelHash,
		Height:             chain.Height,
		Timestamp:          chain.Timestamp,
		Target:             chain.Target,
		BigTotalDifficulty: chain.BigTotalDifficulty,
		Transactions:       chain.Transactions,
		forging:            chain.forging,
		mempool:            chain.mempool,
	}
	mainChainBigTotalDifficulty := chain.BigTotalDifficulty

	var accs *accounts.Accounts
	var toks *tokens.Tokens

	boltTx, err := store.StoreBlockchain.DB.Begin(true)
	if err != nil {
		return
	}

	insertedBlocks := make([]*block.BlockComplete, 0)
	insertedTxHashes := make([][]byte, 0)

	var writer *bolt.Bucket
	savedBlock := false
	//remove blocks which are different
	removedTxHashes := make(map[string][]byte)
	removedTx := make([][]byte, 0)
	removedBlocksHeights := make([]uint64, 0)

	func() {

		defer func() {

			_ = helpers.ConvertRecoverError(recover())

			//recover, but in case the chain was correctly saved and the mewChainDifficulty is higher than
			//we should store it
			if savedBlock && mainChainBigTotalDifficulty.Cmp(newChain.BigTotalDifficulty) < 0 {

				newChain.saveBlockchain(writer)

				for _, removedBlock := range removedBlocksHeights {
					newChain.deleteUnusedBlocksComplete(writer, removedBlock, accs, toks)
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
					chain.Height = newChain.Height
					chain.Hash = newChain.Hash
					chain.PrevHash = newChain.PrevHash
					chain.KernelHash = newChain.KernelHash
					chain.PrevKernelHash = newChain.PrevKernelHash
					chain.Timestamp = newChain.Timestamp
					chain.Target = newChain.Target
					chain.Transactions = newChain.Transactions
					chain.BigTotalDifficulty = newChain.BigTotalDifficulty
				}

				chain.Unlock()

			} else {

				err = boltTx.Rollback()
				if err == nil {
					err = errors.New("Blocks were not saved")
				}

			}

		}()

		writer = boltTx.Bucket([]byte("Chain"))

		accs = accounts.NewAccounts(boltTx)
		toks = tokens.NewTokens(boltTx)

		//let's filter existing blocks
		for i := len(blocksComplete) - 1; i >= 0; i-- {

			blkComplete := blocksComplete[i]

			if blkComplete.Block.Height < newChain.Height {
				hash := newChain.loadBlockHash(writer, blkComplete.Block.Height)
				hash2 := blkComplete.Block.ComputeHash()
				if bytes.Equal(hash, hash2) {
					blocksComplete = blocksComplete[i+1:]
					break
				}
			}

		}

		if len(blocksComplete) == 0 {
			panic("blocks are identical now")
		}

		firstBlockComplete := blocksComplete[0]
		if firstBlockComplete.Block.Height < newChain.Height {
			for i := newChain.Height - 1; i >= newChain.Height; i-- {
				removedBlocksHeights = append(removedBlocksHeights, 0)
				copy(removedBlocksHeights[1:], removedBlocksHeights)
				removedBlocksHeights[0] = i

				newChain.removeBlockComplete(writer, i, removedTxHashes, accs, toks)
			}
		}

		if blocksComplete[0].Block.Height != newChain.Height {
			panic("First Block has is not matching")
		}

		if !bytes.Equal(firstBlockComplete.Block.PrevHash, newChain.Hash) {
			panic("First block hash is not matching chain hash")
		}

		if !bytes.Equal(firstBlockComplete.Block.PrevKernelHash, newChain.KernelHash) {
			panic("First block kernel hash is not matching chain prev kerneh lash")
		}

		for i, blkComplete := range blocksComplete {

			//check block height
			if blkComplete.Block.Height != newChain.Height {
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

				if !bytes.Equal(blkComplete.Block.DelegatedPublicKey, acc.DelegatedStake.DelegatedPublicKey) {
					panic("Block Staking Delegated Public Key is not matching")
				}

			}

			if blkComplete.Block.StakingAmount > stakingAmount {
				panic("Block Staking Amount doesn't match")
			}

			if blkComplete.Block.StakingAmount < stake.GetRequiredStake(blkComplete.Block.Height) {
				panic("Delegated stake ready amount is not enought")
			}

			hash := blkComplete.Block.ComputeHash()
			kernelHash := blkComplete.Block.ComputeKernelHash()

			if difficulty.CheckKernelHashBig(kernelHash, newChain.Target) != true {
				panic("KernelHash Difficulty is not met")
			}

			//already verified for i == 0
			if i > 0 {
				if !bytes.Equal(blkComplete.Block.PrevHash, newChain.Hash) {
					panic("PrevHash doesn't match Genesis prevHash")
				}
				if !bytes.Equal(blkComplete.Block.PrevKernelHash, newChain.KernelHash) {
					panic("PrevHash doesn't match Genesis prevKernelHash")
				}
			}

			blkComplete.Validate()
			blkComplete.Verify()

			if blkComplete.Block.Timestamp < newChain.Timestamp {
				panic("Timestamp has to be greather than the last timestmap")
			}

			if blkComplete.Block.Timestamp > uint64(time.Now().UTC().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
				panic("Timestamp is too much into the future")
			}

			blkComplete.IncludeBlockComplete(accs, toks)

			//to detect if the savedBlock was done correctly
			savedBlock = false

			newTransactionsSaved := newChain.saveBlockComplete(writer, blkComplete, hash, removedTxHashes, accs, toks)

			if len(removedBlocksHeights) > 0 {
				removedBlocksHeights = removedBlocksHeights[1:]
			}

			accs.Commit() //it will commit the changes but not save them
			toks.Commit() //it will commit the changes but not save them

			newChain.PrevHash = newChain.Hash
			newChain.Hash = hash
			newChain.PrevKernelHash = newChain.KernelHash
			newChain.KernelHash = kernelHash
			newChain.Timestamp = blkComplete.Block.Timestamp

			difficultyBigInt := difficulty.ConvertTargetToDifficulty(newChain.Target)
			newChain.BigTotalDifficulty = new(big.Int).Add(newChain.BigTotalDifficulty, difficultyBigInt)
			newChain.saveTotalDifficultyExtra(writer)

			newChain.Target = newChain.computeNextTargetBig(writer)

			newChain.Height += 1
			newChain.Transactions += uint64(len(blkComplete.Txs))
			insertedBlocks = append(insertedBlocks, blkComplete)

			for _, txHashId := range newTransactionsSaved {
				insertedTxHashes = append(insertedTxHashes, txHashId)
			}

			writer.Put([]byte("chainHash"), newChain.Hash)
			writer.Put([]byte("chainPrevHash"), newChain.PrevHash)
			writer.Put([]byte("chainKernelHash"), newChain.KernelHash)
			writer.Put([]byte("chainPrevKernelHash"), newChain.PrevKernelHash)

			buf := make([]byte, binary.MaxVarintLen64)
			n := binary.PutUvarint(buf, newChain.Height)
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
	gui.Warning(fmt.Sprintf("Including blocks SUCCESS %s", hex.EncodeToString(chain.Hash)))
	gui.Warning("-------------------------------------------")
	newChain.updateChainInfo()

	chain.UpdateChannel <- newChain.Height //sending 1
	chain.UpdateNewChainChannel <- &newChain

	//accs will only be read only
	newChain.forging.Wallet.UpdateBalanceChanges(accs)

	//create next block and the workers will be automatically reset
	newChain.createNextBlockForForging()

	for _, txData := range removedTx {
		tx := transaction.Transaction{}
		tx.Deserialize(helpers.NewBufferReader(txData))
		newChain.mempool.AddTxToMemPoolSilent(&tx, newChain.Height, false)
	}

	for _, txHash := range insertedTxHashes {
		newChain.mempool.Delete(txHash)
	}

	newChain.mempool.UpdateChanges(newChain.Hash, newChain.Height)

	result = true
	return

}

func BlockchainInit(forging *forging.Forging, mempool *mempool.MemPool) (chain *Blockchain) {

	gui.Log("Blockchain init...")

	genesis.GenesisInit()

	chain = &Blockchain{
		forging:               forging,
		mempool:               mempool,
		Sync:                  false,
		UpdateChannel:         make(chan uint64),
		UpdateNewChainChannel: make(chan *Blockchain),
	}

	success, err := chain.loadBlockchain()
	if err != nil {
		panic(err)
	}

	if !success {
		chain.init()
	}

	chain.updateChainInfo()
	chain.initForging()

	return
}

func (chain *Blockchain) initForging() {

	go func() {

		for {

			blkComplete := <-chain.forging.SolutionChannel

			var array []*block.BlockComplete
			array = append(array, blkComplete)

			result, err := chain.AddBlocks(array, true)
			if err == nil && result {
				gui.Info("Block was forged! " + strconv.FormatUint(blkComplete.Block.Height, 10))
			} else {
				gui.Error("Error forging block "+strconv.FormatUint(blkComplete.Block.Height, 10), err)
			}

		}

	}()

	go chain.createNextBlockForForging()

}

func (chain *Blockchain) Close() {
}
