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
	"pandora-pay/blockchain/accounts/account"
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
	"pandora-pay/wallet"
	"sync"
	"sync/atomic"
	"time"
)

type Blockchain struct {
	ChainData               *atomic.Value //*BlockchainData
	Sync                    *BlockchainSync
	forging                 *forging.Forging          `json:"-"`
	mempool                 *mempool.Mempool          `json:"-"`
	wallet                  *wallet.Wallet            `json:"-"`
	mutex                   *sync.Mutex               `json:"-"` //writing mutex
	UpdateMulticast         *helpers.MulticastChannel `json:"-"` //chan uint64
	UpdateNewChainMulticast *helpers.MulticastChannel `json:"-"` //chan *BlockchainData
}

func (chain *Blockchain) validateBlocks(blocksComplete []*block_complete.BlockComplete) (err error) {
	if len(blocksComplete) == 0 {
		return errors.New("Blocks length is ZERO")
	}

	for _, blkComplete := range blocksComplete {
		if err = blkComplete.Validate(); err != nil {
			return
		}
		if err = blkComplete.Verify(); err != nil {
			return
		}
	}

	return
}

func (chain *Blockchain) AddBlocks(blocksComplete []*block_complete.BlockComplete, calledByForging bool) (err error) {

	if err = chain.validateBlocks(blocksComplete); err != nil {
		return
	}

	//avoid processing the same function twice
	chain.mutex.Lock()

	chainData := chain.GetChainData()

	gui.Info(fmt.Sprintf("Including blocks %d ... %d", chainData.Height, chainData.Height+uint64(len(blocksComplete))))

	//chain.RLock() is not required because it is guaranteed that no other thread is writing now in the chain
	var newChainData = &BlockchainData{
		Hash:                  chainData.Hash,
		PrevHash:              chainData.PrevHash,
		KernelHash:            chainData.KernelHash,
		PrevKernelHash:        chainData.PrevKernelHash,
		Height:                chainData.Height,
		Timestamp:             chainData.Timestamp,
		Target:                chainData.Target,
		BigTotalDifficulty:    chainData.BigTotalDifficulty,
		ConsecutiveSelfForged: chainData.ConsecutiveSelfForged,
		Transactions:          chainData.Transactions,
	}

	insertedBlocks := []*block_complete.BlockComplete{}
	insertedTxHashes := [][]byte{}

	//remove blocks which are different
	removedTxHashes := make(map[string][]byte)
	removedTx := [][]byte{}
	removedBlocksHeights := []uint64{}

	var accs *accounts.Accounts
	var toks *tokens.Tokens

	err = func() (err error) {

		var boltTx *bolt.Tx
		boltTxClosed := false
		if boltTx, err = store.StoreBlockchain.DB.Begin(true); err != nil {
			return
		}
		defer func() {
			if !boltTxClosed {
				err2 := boltTx.Rollback()
				if err == nil {
					err = err2
				}
			}
		}()

		var writer *bolt.Bucket
		savedBlock := false

		writer = boltTx.Bucket([]byte("Chain"))

		accs = accounts.NewAccounts(boltTx)
		toks = tokens.NewTokens(boltTx)

		err = func() (err error) {
			//let's filter existing blocks
			for i := len(blocksComplete) - 1; i >= 0; i-- {

				blkComplete := blocksComplete[i]

				if blkComplete.Block.Height < newChainData.Height {
					var hash []byte
					if hash, err = chain.LoadBlockHash(writer, blkComplete.Block.Height); err != nil {
						return
					}
					if bytes.Equal(hash, blkComplete.Block.Bloom.Hash) {
						blocksComplete = blocksComplete[i+1:]
						break
					}
				}

			}

			if len(blocksComplete) == 0 {
				return errors.New("blocks are identical now")
			}

			firstBlockComplete := blocksComplete[0]
			if firstBlockComplete.Block.Height < newChainData.Height {

				index := newChainData.Height - 1
				for {

					removedBlocksHeights = append(removedBlocksHeights, 0)
					copy(removedBlocksHeights[1:], removedBlocksHeights)
					removedBlocksHeights[0] = index

					if err = chain.removeBlockComplete(writer, index, removedTxHashes, accs, toks); err != nil {
						return
					}

					if index > firstBlockComplete.Block.Height {
						index -= 1
					} else {
						break
					}
				}

				if firstBlockComplete.Block.Height == 0 {
					newChainData = chain.createGenesisBlockchainData()
				} else {
					if err = newChainData.loadBlockchainInfo(writer, firstBlockComplete.Block.Height); err != nil {
						return
					}
				}
			}

			if blocksComplete[0].Block.Height != newChainData.Height {
				return errors.New("First block hash is not matching")
			}

			if !bytes.Equal(firstBlockComplete.Block.PrevHash, newChainData.Hash) {
				return errors.New("First block hash is not matching chain hash")
			}

			if !bytes.Equal(firstBlockComplete.Block.PrevKernelHash, newChainData.KernelHash) {
				return errors.New("First block kernel hash is not matching chain prev kerneh lash")
			}

			for i, blkComplete := range blocksComplete {

				//check block height
				if blkComplete.Block.Height != newChainData.Height {
					return errors.New("Block Height is not right!")
				}

				//check blkComplete balance

				var acc *account.Account
				if acc, err = accs.GetAccount(blkComplete.Block.Forger, blkComplete.Block.Height); err != nil {
					return
				}

				if acc == nil || !acc.HasDelegatedStake() {
					return errors.New("Forger Account deson't exist or hasn't delegated stake")
				}

				stakingAmount := acc.GetDelegatedStakeAvailable()

				if !bytes.Equal(blkComplete.Block.Bloom.DelegatedPublicKeyHash, acc.DelegatedStake.DelegatedPublicKeyHash) {
					return errors.New("Block Staking Delegated Public Key is not matching")
				}

				if blkComplete.Block.StakingAmount != stakingAmount {
					return errors.New("Block Staking Amount doesn't match")
				}

				if blkComplete.Block.StakingAmount < stake.GetRequiredStake(blkComplete.Block.Height) {
					return errors.New("Delegated stake ready amount is not enought")
				}

				//gui.Log("Staking amount ", newChainData.Height, "value", blkComplete.Block.StakingAmount)
				//gui.Log("Target check ", newChainData.Height, "value", newChainData.Target.Text(10))

				if difficulty.CheckKernelHashBig(blkComplete.Block.Bloom.KernelHash, newChainData.Target) != true {
					return errors.New("KernelHash Difficulty is not met")
				}

				//already verified for i == 0
				if i > 0 {
					if !bytes.Equal(blkComplete.Block.PrevHash, newChainData.Hash) {
						return errors.New("PrevHash doesn't match Genesis prevHash")
					}
					if !bytes.Equal(blkComplete.Block.PrevKernelHash, newChainData.KernelHash) {
						return errors.New("PrevHash doesn't match Genesis prevKernelHash")
					}
				}

				if blkComplete.Block.Timestamp < newChainData.Timestamp {
					return errors.New("Timestamp has to be greather than the last timestmap")
				}

				if blkComplete.Block.Timestamp > uint64(time.Now().UTC().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
					return errors.New("Timestamp is too much into the future")
				}

				if err = blkComplete.IncludeBlockComplete(accs, toks); err != nil {
					return errors.New("Error including block into Blockchain: " + err.Error())
				}

				//to detect if the savedBlock was done correctly
				savedBlock = false

				var newTransactionsSaved [][]byte
				if newTransactionsSaved, err = chain.saveBlockComplete(writer, blkComplete, blkComplete.Block.Bloom.Hash, removedTxHashes, accs, toks); err != nil {
					return
				}

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
				if err = newChainData.saveTotalDifficultyExtra(writer); err != nil {
					return
				}

				if newChainData.Target, err = newChainData.computeNextTargetBig(writer); err != nil {
					return
				}

				//gui.Log("Target new   ", newChainData.Height, "value", newChainData.Target.Text(10))

				newChainData.Height += 1
				newChainData.Transactions += uint64(len(blkComplete.Txs))
				insertedBlocks = append(insertedBlocks, blkComplete)

				for _, txHashId := range newTransactionsSaved {
					insertedTxHashes = append(insertedTxHashes, txHashId)
				}

				if err = writer.Put([]byte("chainHash"), newChainData.Hash); err != nil {
					return
				}
				if err = writer.Put([]byte("chainPrevHash"), newChainData.PrevHash); err != nil {
					return
				}
				if err = writer.Put([]byte("chainKernelHash"), newChainData.KernelHash); err != nil {
					return
				}
				if err = writer.Put([]byte("chainPrevKernelHash"), newChainData.PrevKernelHash); err != nil {
					return
				}

				buf := make([]byte, binary.MaxVarintLen64)
				n := binary.PutUvarint(buf, newChainData.Height)
				if err = writer.Put([]byte("chainHeight"), buf[:n]); err != nil {
					return
				}

				if err = newChainData.saveBlockchainInfo(writer); err != nil {
					return
				}

				savedBlock = true
			}

			return
		}()

		//recover, but in case the chain was correctly saved and the mewChainDifficulty is higher than
		//we should store it
		if savedBlock && chainData.BigTotalDifficulty.Cmp(newChainData.BigTotalDifficulty) < 0 {

			if calledByForging {
				newChainData.ConsecutiveSelfForged += 1
			} else {
				newChainData.ConsecutiveSelfForged = 0
			}

			if err = newChainData.saveBlockchain(writer); err != nil {
				panic("Error saving Blockchain " + err.Error())
			}

			for _, removedBlock := range removedBlocksHeights {
				if err = chain.deleteUnusedBlocksComplete(writer, removedBlock, accs, toks); err != nil {
					panic("Error deleting unused blocks Blockchain " + err.Error())
				}
			}
			for txHash := range removedTxHashes {
				data := writer.Get(append([]byte("tx"), txHash...))
				removedTx = append(removedTx, data)
				if err = writer.Delete(append([]byte("tx"), txHash...)); err != nil {
					panic("Error deleting transactions " + err.Error())
				}
			}

			accs.Rollback()
			toks.Rollback()
			if err = accs.WriteToStore(); err != nil {
				panic("Error writing accs" + err.Error())
			}
			if err = toks.WriteToStore(); err != nil {
				panic("Error writing accs" + err.Error())
			}

			if err = boltTx.Commit(); err != nil {
				panic("Error storing writing changes to disk" + err.Error())
			}
			chain.ChainData.Store(newChainData)

			boltTxClosed = true

		} else {
			//only rollback
		}

		return
	}()

	if err == nil && len(insertedBlocks) == 0 {
		err = errors.New("No blocks were inserted")
	}

	if err != nil {
		if calledByForging {
			chain.createNextBlockForForging()
		}
		chain.mutex.Unlock()
		return
	}

	gui.Warning("-------------------------------------------")
	gui.Warning(fmt.Sprintf("Including blocks %d | TXs: %d | Hash %s", len(insertedBlocks), len(insertedTxHashes), hex.EncodeToString(chainData.Hash)))
	gui.Warning("-------------------------------------------")
	newChainData.updateChainInfo()

	//accs will only be read only
	if err = chain.forging.Wallet.UpdateAccountsChanges(accs); err != nil {
		gui.Error("Error updating balance changes", err)
	}

	if err = chain.wallet.UpdateAccountsChanges(accs); err != nil {
		gui.Error("Error updating balance changes", err)
	}

	if err = chain.forging.Wallet.ProcessUpdates(); err != nil {
		gui.Error("Error Processing Updates", err)
	}

	chain.mutex.Unlock()

	//update work for mem pool
	chain.mempool.UpdateWork(newChainData.Hash, newChainData.Height)

	//create next block and the workers will be automatically reset
	chain.createNextBlockForForging()

	for _, txData := range removedTx {
		tx := &transaction.Transaction{}
		if err = tx.Deserialize(helpers.NewBufferReader(txData)); err != nil {
			return
		}
		if err = tx.BloomExtraNow(true); err != nil {
			return
		}
		if _, err = chain.mempool.AddTxToMemPool(tx, newChainData.Height, false); err != nil {
			return
		}
	}

	chain.mempool.DeleteTxs(insertedTxHashes)

	newSyncTime, result := chain.Sync.addBlocksChanged(uint32(len(insertedBlocks)), false)

	chain.UpdateMulticast.Broadcast(newChainData.Height)

	chain.UpdateNewChainMulticast.Broadcast(newChainData)

	if result {
		chain.Sync.UpdateSyncMulticast.Broadcast(newSyncTime)
	}

	return
}

func BlockchainInit(forging *forging.Forging, wallet *wallet.Wallet, mempool *mempool.Mempool) (chain *Blockchain, err error) {

	gui.Log("Blockchain init...")

	if err = genesis.GenesisInit(wallet); err != nil {
		return
	}

	chain = &Blockchain{
		ChainData:               &atomic.Value{},
		mutex:                   &sync.Mutex{},
		forging:                 forging,
		mempool:                 mempool,
		wallet:                  wallet,
		Sync:                    createBlockchainSync(),
		UpdateMulticast:         helpers.NewMulticastChannel(),
		UpdateNewChainMulticast: helpers.NewMulticastChannel(),
	}

	if err = chain.loadBlockchain(); err != nil {
		if err.Error() != "Chain not found" {
			return
		}
		if _, err = chain.init(); err != nil {
			return
		}
		if err = chain.saveBlockchain(); err != nil {
			return
		}
	}

	chainData := chain.GetChainData()
	chainData.updateChainInfo()

	if err = wallet.ReadWallet(); err != nil {
		return
	}
	chain.initForging()

	return
}

func (chain *Blockchain) Close() {
	chain.UpdateNewChainMulticast.CloseAll()
	chain.UpdateMulticast.CloseAll()
}
