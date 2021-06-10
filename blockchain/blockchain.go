package blockchain

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/blocks/block-complete"
	difficulty "pandora-pay/blockchain/blocks/block/difficulty"
	"pandora-pay/blockchain/forging/forging-block-work"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/config"
	"pandora-pay/config/stake"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"pandora-pay/wallet"
	"sync"
	"sync/atomic"
	"time"
)

type Blockchain struct {
	ChainData                *atomic.Value //*BlockchainData
	Sync                     *BlockchainSync
	mempool                  *mempool.Mempool
	wallet                   *wallet.Wallet
	mutex                    *sync.Mutex //writing mutex
	updatesQueue             *BlockchainUpdatesQueue
	ForgingSolutionCn        chan *block_complete.BlockComplete
	UpdateNewChain           *multicast.MulticastChannel          //chan uint64
	UpdateNewChainDataUpdate *multicast.MulticastChannel          //chan *BlockchainDataUpdate
	UpdateAccounts           *multicast.MulticastChannel          //chan *accounts
	UpdateTokens             *multicast.MulticastChannel          //chan *tokens
	NextBlockCreatedCn       chan *forging_block_work.ForgingWork //chan
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

	gui.GUI.Info(fmt.Sprintf("Including blocks %d ... %d", chainData.Height, chainData.Height+uint64(len(blocksComplete))))

	//chain.RLock() is not required because it is guaranteed that no other thread is writing now in the chain
	var newChainData = &BlockchainData{
		Hash:                  helpers.CloneBytes(chainData.Hash),             //atomic copy
		PrevHash:              helpers.CloneBytes(chainData.PrevHash),         //atomic copy
		KernelHash:            helpers.CloneBytes(chainData.KernelHash),       //atomic copy
		PrevKernelHash:        helpers.CloneBytes(chainData.PrevKernelHash),   //atomic copy
		Height:                chainData.Height,                               //atomic copy
		Timestamp:             chainData.Timestamp,                            //atomic copy
		Target:                new(big.Int).Set(chainData.Target),             //atomic copy
		BigTotalDifficulty:    new(big.Int).Set(chainData.BigTotalDifficulty), //atomic copy
		ConsecutiveSelfForged: chainData.ConsecutiveSelfForged,                //atomic copy
		TransactionsCount:     chainData.TransactionsCount,                    //atomic copy
	}

	insertedBlocks := []*block_complete.BlockComplete{}
	insertedTxHashes := [][]byte{}

	//remove blocks which are different
	removedTxHashes := make(map[string][]byte)
	removedTxs := [][]byte{}
	removedBlocksHeights := []uint64{}
	removedBlocksTransactionsCount := uint64(0)

	var accs *accounts.Accounts
	var toks *tokens.Tokens

	err = func() (err error) {

		chain.mempool.SuspendProcessingCn <- struct{}{}

		err = store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

			savedBlock := false

			accs = accounts.NewAccounts(writer)
			toks = tokens.NewTokens(writer)

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
					removedBlocksTransactionsCount = 0
				} else {
					removedBlocksTransactionsCount = newChainData.TransactionsCount
					newChainData = &BlockchainData{}
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

			err = func() (err error) {

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
					if newTransactionsSaved, err = chain.saveBlockComplete(writer, blkComplete, newChainData.TransactionsCount, removedTxHashes, accs, toks); err != nil {
						return errors.New("Error saving block complete: " + err.Error())
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

					if newChainData.Target, err = newChainData.computeNextTargetBig(writer); err != nil {
						return
					}

					newChainData.Height += 1
					newChainData.TransactionsCount += uint64(len(blkComplete.Txs))
					insertedBlocks = append(insertedBlocks, blkComplete)

					for _, txHashId := range newTransactionsSaved {
						insertedTxHashes = append(insertedTxHashes, txHashId)
					}

					if err = newChainData.saveTotalDifficultyExtra(writer); err != nil {
						return
					}

					if err = writer.Put("chainHash", newChainData.Hash); err != nil {
						return
					}
					if err = writer.Put("chainPrevHash", newChainData.PrevHash); err != nil {
						return
					}
					if err = writer.Put("chainKernelHash", newChainData.KernelHash); err != nil {
						return
					}
					if err = writer.Put("chainPrevKernelHash", newChainData.PrevKernelHash); err != nil {
						return
					}

					if err = newChainData.saveBlockchainHeight(writer); err != nil {
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

				if len(removedBlocksHeights) > 0 {

					//remove unused blocks
					for _, removedBlock := range removedBlocksHeights {
						if err = chain.deleteUnusedBlocksComplete(writer, removedBlock, accs, toks); err != nil {
							panic("Error deleting unused blocks: " + err.Error())
						}
					}

					//removing unused transactions
					if config.SEED_WALLET_NODES_INFO {
						if err = removeUnusedTransactions(writer, newChainData.TransactionsCount, removedBlocksTransactionsCount); err != nil {
							panic(err)
						}
					}
				}

				for txHash := range removedTxHashes {
					removedTxs = append(removedTxs, writer.GetClone("tx"+txHash)) //required because the garbage collector sometimes it deletes the underlying buffers
					if err = writer.Delete("tx" + txHash); err != nil {
						panic("Error deleting transaction: " + err.Error())
					}
				}

				if config.SEED_WALLET_NODES_INFO {
					if err = removeTxsInfo(writer, removedTxHashes); err != nil {
						panic(err)
					}
				}

				if err = chain.saveBlockchainHashmaps(accs, toks); err != nil {
					panic(err)
				}

				chain.ChainData.Store(newChainData)

			} else {
				//only rollback
				if err == nil {
					err = errors.New("Rollback")
				}
			}

			if accs != nil {
				accs.UnsetTx()
				toks.UnsetTx()
			}

			return
		})

		return
	}()

	if err == nil && len(insertedBlocks) == 0 {
		err = errors.New("No blocks were inserted")
	}

	update := &BlockchainUpdate{
		err:             err,
		calledByForging: calledByForging,
	}

	if err == nil {
		update.newChainData = newChainData
		update.accs = accs
		update.toks = toks
		update.removedTxs = removedTxs
		update.insertedBlocks = insertedBlocks
		update.insertedTxHashes = insertedTxHashes
	}

	gui.GUI.Log("writing to chain.updatesQueue.updatesCn")

	chain.updatesQueue.updatesCn <- update

	chain.mutex.Unlock()

	return
}

func CreateBlockchain(mempool *mempool.Mempool) (chain *Blockchain, err error) {

	gui.GUI.Log("Blockchain init...")

	chain = &Blockchain{
		ChainData:                &atomic.Value{},
		mutex:                    &sync.Mutex{},
		mempool:                  mempool,
		updatesQueue:             createBlockchainUpdatesQueue(),
		Sync:                     createBlockchainSync(),
		ForgingSolutionCn:        make(chan *block_complete.BlockComplete),
		UpdateNewChain:           multicast.NewMulticastChannel(),
		UpdateNewChainDataUpdate: multicast.NewMulticastChannel(),
		UpdateAccounts:           multicast.NewMulticastChannel(),
		UpdateTokens:             multicast.NewMulticastChannel(),
		NextBlockCreatedCn:       make(chan *forging_block_work.ForgingWork),
	}

	chain.updatesQueue.chain = chain
	chain.updatesQueue.processQueue()

	return
}

func (chain *Blockchain) InitializeChain() (err error) {

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

	chain.InitForging()

	return
}

func (chain *Blockchain) Close() {
	chain.UpdateNewChainDataUpdate.CloseAll()
	chain.UpdateNewChain.CloseAll()
	chain.UpdateAccounts.CloseAll()
	chain.UpdateTokens.CloseAll()
	close(chain.NextBlockCreatedCn)
	close(chain.ForgingSolutionCn)
}
