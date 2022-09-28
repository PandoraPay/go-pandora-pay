package blockchain

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block/difficulty"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
	"strconv"
	"sync"
	"time"
)

type Blockchain struct {
	ChainData                               *generics.Value[*BlockchainData]
	Sync                                    *blockchain_sync.BlockchainSync
	mempool                                 *mempool.Mempool
	wallet                                  *wallet.Wallet
	txsValidator                            *txs_validator.TxsValidator
	mutex                                   *sync.Mutex //writing mutex
	updatesQueue                            *BlockchainUpdatesQueue
	ForgingSolutionCn                       chan *blockchain_types.BlockchainSolution
	UpdateNewChain                          *multicast.MulticastChannel[uint64]
	UpdateNewChainDataUpdate                *multicast.MulticastChannel[*BlockchainDataUpdate]
	UpdateNewChainUpdate                    *multicast.MulticastChannel[*blockchain_types.BlockchainUpdates]
	UpdateSocketsSubscriptionsTransactions  *multicast.MulticastChannel[[]*blockchain_types.BlockchainTransactionUpdate]
	UpdateSocketsSubscriptionsNotifications *multicast.MulticastChannel[*data_storage.DataStorage]
	NextBlockCreatedCn                      chan *forging_block_work.ForgingWork
}

func (chain *Blockchain) validateBlocks(blocksComplete []*block_complete.BlockComplete) (err error) {

	if len(blocksComplete) == 0 {
		return errors.New("Blocks length is ZERO")
	}

	for _, blkComplete := range blocksComplete {

		if err = blkComplete.Verify(); err != nil {
			return
		}

		if err = chain.txsValidator.ValidateTxs(blkComplete.Txs); err != nil {
			return
		}

	}

	return
}

func (chain *Blockchain) AddBlocks(blocksComplete []*block_complete.BlockComplete, calledByForging bool, exceptSocketUUID advanced_connection_types.UUID) (kernelHash []byte, err error) {

	if err = chain.validateBlocks(blocksComplete); err != nil {
		return
	}

	//avoid processing the same function twice
	chain.mutex.Lock()
	defer chain.mutex.Unlock()

	chainData := chain.GetChainData()

	if calledByForging && blocksComplete[len(blocksComplete)-1].Height == chainData.Height-1 && chainData.ConsecutiveSelfForged > 0 {
		err = errors.New("Block was already forged by a different thread")
		return
	}

	gui.GUI.Info("Including blocks " + strconv.FormatUint(chainData.Height, 10) + " ... " + strconv.FormatUint(chainData.Height+uint64(len(blocksComplete)), 10))

	//chain.RLock() is not required because it is guaranteed that no other thread is writing now in the chain
	var newChainData = &BlockchainData{
		helpers.CloneBytes(chainData.Hash),             //atomic copy
		helpers.CloneBytes(chainData.PrevHash),         //atomic copy
		helpers.CloneBytes(chainData.KernelHash),       //atomic copy
		helpers.CloneBytes(chainData.PrevKernelHash),   //atomic copy
		chainData.Height,                               //atomic copy
		chainData.Timestamp,                            //atomic copy
		new(big.Int).Set(chainData.Target),             //atomic copy
		new(big.Int).Set(chainData.BigTotalDifficulty), //atomic copy
		chainData.TransactionsCount,                    //atomic copy
		chainData.AccountsCount,                        //atomic copy
		chainData.AssetsCount,                          //atomic copy
		chainData.Supply,
		chainData.ConsecutiveSelfForged, //atomic copy
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

	var dataStorage *data_storage.DataStorage

	err = func() (err error) {

		chain.mempool.SuspendProcessingCn <- struct{}{}

		err = store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

			defer func() {
				if errReturned := recover(); errReturned != nil {
					err = errReturned.(error)
				}
			}()

			savedBlock := false

			dataStorage = data_storage.NewDataStorage(writer)

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

					if allTransactionsChanges, err = chain.removeBlockComplete(writer, index, removedTxHashes, allTransactionsChanges, dataStorage); err != nil {
						return
					}

					if index > firstBlockComplete.Block.Height {
						index -= 1
					} else {
						break
					}
				}

				if firstBlockComplete.Block.Height == 0 {
					gui.GUI.Info("chain.createGenesisBlockchainData called")
					newChainData = chain.createGenesisBlockchainData()
					removedBlocksTransactionsCount = 0
				} else {
					removedBlocksTransactionsCount = newChainData.TransactionsCount
					newChainData = &BlockchainData{}
					if err = newChainData.loadBlockchainInfo(writer, firstBlockComplete.Block.Height); err != nil {
						return
					}
				}

				if err = dataStorage.CommitChanges(); err != nil {
					return
				}

			}

			if blocksComplete[0].Block.Height != newChainData.Height {
				return errors.New("First block hash is not matching")
			}

			if !bytes.Equal(firstBlockComplete.Block.PrevHash, newChainData.Hash) {
				return fmt.Errorf("First block hash is not matching chain hash %d %s %s ", firstBlockComplete.Block.Height, base64.StdEncoding.EncodeToString(firstBlockComplete.Bloom.Hash), base64.StdEncoding.EncodeToString(newChainData.Hash))
			}

			if !bytes.Equal(firstBlockComplete.Block.PrevKernelHash, newChainData.KernelHash) {
				return errors.New("First block kernel hash is not matching chain prev kerneh lash")
			}

			err = func() (err error) {

				for _, blkComplete := range blocksComplete {

					//check block height
					if blkComplete.Block.Height != newChainData.Height {
						return errors.New("Block Height is not right!")
					}

					//increase supply
					var ast *asset.Asset
					if ast, err = dataStorage.Asts.GetAsset(config_coins.NATIVE_ASSET_FULL); err != nil {
						return
					}

					var reward, _ uint64
					if reward, _, err = blockchain_types.ComputeBlockReward(blkComplete.Height, blkComplete.Txs); err != nil {
						return
					}

					if err = ast.AddNativeSupply(true, reward); err != nil {
						return
					}
					if err = dataStorage.Asts.Update(string(config_coins.NATIVE_ASSET_FULL), ast); err != nil {
						return
					}
					newChainData.Supply = ast.Supply

					var plainAcc *plain_account.PlainAccount
					if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(blkComplete.Block.Forger); err != nil {
						return
					}

					if plainAcc == nil {
						return errors.New("Forger Account deson't exist or hasn't delegated stake")
					}

					if !bytes.Equal(blkComplete.Block.DelegatedStakePublicKey, plainAcc.DelegatedStake.DelegatedStakePublicKey) {
						return errors.New("Block Staking Delegated Public Key is not matching")
					}

					if blkComplete.Block.DelegatedStakeFee != plainAcc.DelegatedStake.DelegatedStakeFee {
						return fmt.Errorf("Block Delegated Stake Fee doesn't match %d %d", blkComplete.Block.DelegatedStakeFee, plainAcc.DelegatedStake.DelegatedStakeFee)
					}

					if blkComplete.Block.StakingAmount != plainAcc.StakeAvailable {
						return fmt.Errorf("Block Staking Amount doesn't match %d %d", blkComplete.Block.StakingAmount, plainAcc.StakeAvailable)
					}

					if blkComplete.Block.StakingAmount < config_stake.GetRequiredStake(blkComplete.Block.Height) {
						return errors.New("Delegated stake ready amount is not enought")
					}

					if difficulty.CheckKernelHashBig(blkComplete.Block.Bloom.KernelHashStaked, newChainData.Target) != true {
						return errors.New("KernelHash Difficulty is not met")
					}

					if !bytes.Equal(blkComplete.Block.PrevHash, newChainData.Hash) {
						return errors.New("PrevHash doesn't match Genesis prevHash")
					}

					if !bytes.Equal(blkComplete.Block.PrevKernelHash, newChainData.KernelHash) {
						return errors.New("PrevHash doesn't match Genesis prevKernelHash")
					}

					if blkComplete.Block.Timestamp < newChainData.Timestamp {
						return errors.New("Timestamp has to be greater than the last timestmap")
					}

					if blkComplete.Block.Timestamp > uint64(time.Now().UTC().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
						return errors.New("Timestamp is too much into the future")
					}

					if err = blkComplete.IncludeBlockComplete(dataStorage); err != nil {
						return fmt.Errorf("Error including block %d into Blockchain: %s", blkComplete.Height, err.Error())
					}

					if err = dataStorage.ProcessPendingStakes(blkComplete.Height); err != nil {
						return errors.New("Error Processing Pending Stakes: " + err.Error())
					}

					//to detect if the savedBlock was done correctly
					savedBlock = false

					if allTransactionsChanges, err = chain.saveBlockComplete(writer, blkComplete, newChainData.TransactionsCount, removedTxHashes, allTransactionsChanges, dataStorage); err != nil {
						return errors.New("Error saving block complete: " + err.Error())
					}

					if len(removedBlocksHeights) > 0 {
						removedBlocksHeights = removedBlocksHeights[1:]
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

					newChainData.saveTotalDifficultyExtra(writer)

					newChainData.saveBlockchainHeight(writer)
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
					}
				}
				for _, change := range allTransactionsChanges {
					if change.Inserted {
						insertedTxs[change.TxHashStr] = change.Tx
						delete(removedTxHashes, change.TxHashStr)
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
						if err = chain.deleteUnusedBlocksComplete(writer, removedBlock, dataStorage); err != nil {
							panic(err)
						}
					}

					//removing unused transactions
					if config.SEED_WALLET_NODES_INFO {
						removeUnusedTransactions(writer, newChainData.TransactionsCount, removedBlocksTransactionsCount)
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
						removedTxsList[removedCount] = writer.Get("tx:" + change.TxHashStr) //required because the garbage collector sometimes it deletes the underlying buffers
						writer.Delete("tx:" + change.TxHashStr)
						writer.Delete("txHash:" + change.TxHashStr)
						writer.Delete("txBlock:" + change.TxHashStr)
						removedCount += 1
					}
					if change.Inserted && insertedTxs[change.TxHashStr] != nil && removedTxHashes[change.TxHashStr] == nil {
						insertedTxsList[insertedCount] = change.Tx
						insertedCount += 1
					}
				}

				if config.SEED_WALLET_NODES_INFO {
					removeTxsInfo(writer, removedTxHashes)
				}

				if err = chain.saveBlockchainHashmaps(dataStorage); err != nil {
					panic(err)
				}

				newChainData.AssetsCount = dataStorage.Asts.Count
				newChainData.AccountsCount = dataStorage.PlainAccs.Count

			} else if err == nil { //only rollback
				err = errors.New("Rollback")
			}

			if dataStorage != nil {
				dataStorage.SetTx(nil)
			}

			return
		})

		return
	}()

	if err == nil && len(insertedBlocks) == 0 {
		err = errors.New("No blocks were inserted")
	}

	if err == nil {
		kernelHash = newChainData.KernelHash
		chain.ChainData.Store(newChainData)
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
		update.dataStorage = dataStorage
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

func CreateBlockchain(mempool *mempool.Mempool, txsValidator *txs_validator.TxsValidator) (*Blockchain, error) {

	gui.GUI.Log("Blockchain init...")

	chain := &Blockchain{
		&generics.Value[*BlockchainData]{},
		blockchain_sync.CreateBlockchainSync(),
		mempool,
		nil,
		txsValidator,
		&sync.Mutex{},
		createBlockchainUpdatesQueue(txsValidator),
		make(chan *blockchain_types.BlockchainSolution),
		multicast.NewMulticastChannel[uint64](),
		multicast.NewMulticastChannel[*BlockchainDataUpdate](),
		multicast.NewMulticastChannel[*blockchain_types.BlockchainUpdates](),
		multicast.NewMulticastChannel[[]*blockchain_types.BlockchainTransactionUpdate](),
		multicast.NewMulticastChannel[*data_storage.DataStorage](),
		make(chan *forging_block_work.ForgingWork),
	}

	chain.updatesQueue.chain = chain
	chain.updatesQueue.processBlockchainUpdatesQueue()
	chain.updatesQueue.processBlockchainUpdateMempool()
	chain.updatesQueue.processBlockchainUpdateNotifications()

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
	close(chain.ForgingSolutionCn)
	close(chain.NextBlockCreatedCn)
}
