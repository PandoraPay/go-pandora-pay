package blockchain

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"
	"pandora-pay/blockchain/blockchain-sync"
	blockchain_types "pandora-pay/blockchain/blockchain-types"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/blocks/block/difficulty"
	"pandora-pay/blockchain/data/accounts"
	plain_accounts "pandora-pay/blockchain/data/plain-accounts"
	plain_account "pandora-pay/blockchain/data/plain-accounts/plain-account"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/data/tokens"
	"pandora-pay/blockchain/forging/forging-block-work"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	transaction_zether "pandora-pay/blockchain/transactions/transaction/transaction-zether"
	"pandora-pay/config"
	"pandora-pay/config/config_stake"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	advanced_connection_types "pandora-pay/network/websocks/connection/advanced-connection-types"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"pandora-pay/wallet"
	"sync"
	"sync/atomic"
	"time"
)

type Blockchain struct {
	ChainData                *atomic.Value //*BlockchainData
	Sync                     *blockchain_sync.BlockchainSync
	mempool                  *mempool.Mempool
	wallet                   *wallet.Wallet
	mutex                    *sync.Mutex //writing mutex
	updatesQueue             *BlockchainUpdatesQueue
	ForgingSolutionCn        chan *block_complete.BlockComplete
	UpdateNewChain           *multicast.MulticastChannel          //uint64
	UpdateNewChainDataUpdate *multicast.MulticastChannel          //*BlockchainDataUpdate
	UpdateAccounts           *multicast.MulticastChannel          //*accounts
	UpdatePlainAccounts      *multicast.MulticastChannel          //*plainAccounts
	UpdateTokens             *multicast.MulticastChannel          //*tokens
	UpdateRegistrations      *multicast.MulticastChannel          //*registrations
	UpdateTransactions       *multicast.MulticastChannel          //[]*blockchain_types.BlockchainTransactionUpdate
	NextBlockCreatedCn       chan *forging_block_work.ForgingWork //
}

func (chain *Blockchain) validateBlocks(blocksComplete []*block_complete.BlockComplete) (err error) {

	if len(blocksComplete) == 0 {
		return errors.New("Blocks length is ZERO")
	}

	nonceMap := make(map[string]bool)

	for _, blkComplete := range blocksComplete {
		if err = blkComplete.Verify(); err != nil {
			return
		}

		for _, tx := range blkComplete.Txs {
			if tx.Version == transaction_type.TX_ZETHER {
				base := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)

				if nonceMap[string(base.Bloom.Nonce1)] || nonceMap[string(base.Bloom.Nonce2)] {
					return errors.New("Zether Nonce exists")
				}
				nonceMap[string(base.Bloom.Nonce1)] = true
				nonceMap[string(base.Bloom.Nonce2)] = true
			}
		}
	}

	return
}

func (chain *Blockchain) AddBlocks(blocksComplete []*block_complete.BlockComplete, calledByForging bool, exceptSocketUUID advanced_connection_types.UUID) (err error) {

	if err = chain.validateBlocks(blocksComplete); err != nil {
		return
	}

	//avoid processing the same function twice
	chain.mutex.Lock()
	defer chain.mutex.Unlock()

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

	allTransactionsChanges := []*blockchain_types.BlockchainTransactionUpdate{}

	insertedBlocks := []*block_complete.BlockComplete{}

	//remove blocks which are different
	removedTxHashes := make(map[string][]byte)
	insertedTxs := make(map[string]*transaction.Transaction)

	var removedTxsList [][]byte                    //ordered list
	var insertedTxsList []*transaction.Transaction //ordered list

	removedBlocksHeights := []uint64{}
	removedBlocksTransactionsCount := uint64(0)

	var accsCollection *accounts.AccountsCollection
	var toks *tokens.Tokens
	var regs *registrations.Registrations
	var plainAccs *plain_accounts.PlainAccounts

	err = func() (err error) {

		chain.mempool.SuspendProcessingCn <- struct{}{}

		err = store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

			savedBlock := false

			accsCollection = accounts.NewAccountsCollection(writer)
			toks = tokens.NewTokens(writer)
			regs = registrations.NewRegistrations(writer)
			plainAccs = plain_accounts.NewPlainAccounts(writer)

			var accs *accounts.Accounts
			if accs, err = accsCollection.GetMap(config.NATIVE_TOKEN); err != nil {
				return
			}

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

					if allTransactionsChanges, err = chain.removeBlockComplete(writer, index, removedTxHashes, allTransactionsChanges, regs, plainAccs, accsCollection, toks); err != nil {
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

					if err != nil {
						return
					}

					var plainAcc *plain_account.PlainAccount
					if plainAcc, err = plainAccs.GetPlainAccount(blkComplete.Block.Forger, blkComplete.Block.Height); err != nil {
						return
					}

					if plainAcc == nil || !plainAcc.HasDelegatedStake() {
						return errors.New("Forger Account deson't exist or hasn't delegated stake")
					}

					var stakingAmount uint64
					stakingAmount, err = plainAcc.ComputeDelegatedStakeAvailable(newChainData.Height)
					if err != nil {
						return
					}

					if !bytes.Equal(blkComplete.Block.DelegatedPublicKey, plainAcc.DelegatedStake.DelegatedPublicKey) {
						return errors.New("Block Staking Delegated Public Key is not matching")
					}

					if blkComplete.Block.StakingAmount != stakingAmount {
						return errors.New("Block Staking Amount doesn't match")
					}

					if blkComplete.Block.StakingAmount < config_stake.GetRequiredStake(blkComplete.Block.Height) {
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
						return errors.New("Timestamp has to be greater than the last timestmap")
					}

					if blkComplete.Block.Timestamp > uint64(time.Now().UTC().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
						return errors.New("Timestamp is too much into the future")
					}

					if err = blkComplete.IncludeBlockComplete(regs, plainAccs, accsCollection, toks); err != nil {
						return errors.New("Error including block into Blockchain: " + err.Error())
					}

					//to detect if the savedBlock was done correctly
					savedBlock = false

					if allTransactionsChanges, err = chain.saveBlockComplete(writer, blkComplete, newChainData.TransactionsCount, removedTxHashes, allTransactionsChanges, regs, accsCollection, toks); err != nil {
						return errors.New("Error saving block complete: " + err.Error())
					}

					if len(removedBlocksHeights) > 0 {
						removedBlocksHeights = removedBlocksHeights[1:]
					}

					//it will commit the changes but not save them
					if err = accs.CommitChanges(); err != nil {
						return
					}
					if err = toks.CommitChanges(); err != nil {
						return
					}
					if err = regs.CommitChanges(); err != nil {
						return
					}
					if err = plainAccs.CommitChanges(); err != nil {
						return
					}

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

				//let's recompute removedTxHashes
				removedTxHashes = make(map[string][]byte)
				for _, change := range allTransactionsChanges {
					if !change.Inserted {
						removedTxHashes[change.TxHashStr] = change.TxHash
					} else {
						insertedTxs[change.Tx.Bloom.HashStr] = change.Tx
					}
				}

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

				//let's keep the order as well
				var removedCount, insertedCount int
				for _, change := range allTransactionsChanges {
					if !change.Inserted && removedTxHashes[change.TxHashStr] != nil && insertedTxs[change.TxHashStr] == nil {
						removedCount += 1
					}
					if change.Inserted && insertedTxs[change.TxHashStr] != nil && removedTxHashes[change.TxHashStr] == nil {
						insertedCount += 1
					}
				}
				removedTxsList = make([][]byte, removedCount)
				insertedTxsList = make([]*transaction.Transaction, insertedCount)
				removedCount, insertedCount = 0, 0

				for _, change := range allTransactionsChanges {
					if !change.Inserted && removedTxHashes[change.TxHashStr] != nil && insertedTxs[change.TxHashStr] == nil {
						removedTxsList[removedCount] = writer.GetClone("tx" + change.TxHashStr) //required because the garbage collector sometimes it deletes the underlying buffers
						if err = writer.Delete("tx" + change.TxHashStr); err != nil {
							panic("Error deleting transaction: " + err.Error())
						}
						removedCount += 1
					}
					if change.Inserted && insertedTxs[change.TxHashStr] != nil && removedTxHashes[change.TxHashStr] == nil {
						insertedTxsList[insertedCount] = change.Tx
						insertedCount += 1
					}
				}

				if config.SEED_WALLET_NODES_INFO {
					if err = removeTxsInfo(writer, removedTxHashes); err != nil {
						panic(err)
					}
				}

				if err = chain.saveBlockchainHashmaps(regs, plainAccs, accsCollection, toks); err != nil {
					panic(err)
				}

				chain.ChainData.Store(newChainData)

			} else {
				//only rollback
				if err == nil {
					err = errors.New("Rollback")
				}
			}

			if accsCollection != nil {
				accsCollection.UnsetTx()
				toks.UnsetTx()
				regs.UnsetTx()
				plainAccs.UnsetTx()
			}

			return
		})

		return
	}()

	if err == nil && len(insertedBlocks) == 0 {
		err = errors.New("No blocks were inserted")
	}

	if err == nil {
		chain.mempool.ContinueProcessingCn <- mempool.CONTINUE_PROCESSING_NO_ERROR
	} else {
		chain.mempool.ContinueProcessingCn <- mempool.CONTINUE_PROCESSING_ERROR
	}

	update := &BlockchainUpdate{
		err:              err,
		calledByForging:  calledByForging,
		exceptSocketUUID: exceptSocketUUID,
	}

	if err == nil {
		update.newChainData = newChainData
		update.accsCollection = accsCollection
		update.toks = toks
		update.regs = regs
		update.plainAccs = plainAccs
		update.removedTxsList = removedTxsList
		update.removedTxHashes = removedTxHashes
		update.insertedTxs = insertedTxs
		update.insertedTxsList = insertedTxsList
		update.insertedBlocks = insertedBlocks
		update.allTransactionsChanges = allTransactionsChanges
	}

	chain.updatesQueue.updatesCn <- update

	return
}

func CreateBlockchain(mempool *mempool.Mempool) (*Blockchain, error) {

	gui.GUI.Log("Blockchain init...")

	chain := &Blockchain{
		ChainData:                &atomic.Value{}, //*BlockchainData
		mutex:                    &sync.Mutex{},
		mempool:                  mempool,
		updatesQueue:             createBlockchainUpdatesQueue(),
		Sync:                     blockchain_sync.CreateBlockchainSync(),
		ForgingSolutionCn:        make(chan *block_complete.BlockComplete),
		UpdateNewChain:           multicast.NewMulticastChannel(),
		UpdateNewChainDataUpdate: multicast.NewMulticastChannel(),
		UpdateAccounts:           multicast.NewMulticastChannel(),
		UpdatePlainAccounts:      multicast.NewMulticastChannel(),
		UpdateTokens:             multicast.NewMulticastChannel(),
		UpdateRegistrations:      multicast.NewMulticastChannel(),
		UpdateTransactions:       multicast.NewMulticastChannel(),
		NextBlockCreatedCn:       make(chan *forging_block_work.ForgingWork),
	}

	chain.updatesQueue.chain = chain
	chain.updatesQueue.processQueue()

	return chain, nil
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

	return
}

func (chain *Blockchain) Close() {
	chain.UpdateNewChainDataUpdate.CloseAll()
	chain.UpdateNewChain.CloseAll()
	chain.UpdateAccounts.CloseAll()
	chain.UpdatePlainAccounts.CloseAll()
	chain.UpdateTokens.CloseAll()
	chain.UpdateRegistrations.CloseAll()
	close(chain.NextBlockCreatedCn)
	close(chain.ForgingSolutionCn)
}
