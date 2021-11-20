package blockchain

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block/difficulty"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_stake"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet"
	"strconv"
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
	UpdateAssets             *multicast.MulticastChannel          //*assets
	UpdateRegistrations      *multicast.MulticastChannel          //*registrations
	UpdateTransactions       *multicast.MulticastChannel          //[]*blockchain_types.BlockchainTransactionUpdate
	NextBlockCreatedCn       chan *forging_block_work.ForgingWork //
}

func (chain *Blockchain) validateBlocks(blocksComplete []*block_complete.BlockComplete) (err error) {

	if len(blocksComplete) == 0 {
		return errors.New("Blocks length is ZERO")
	}

	for _, blkComplete := range blocksComplete {
		if err = blkComplete.Verify(); err != nil {
			return
		}

		nonceMap := make(map[string]bool)
		for _, tx := range blkComplete.Txs {
			if tx.Version == transaction_type.TX_ZETHER {
				base := tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)

				for t := range base.Payloads {
					if nonceMap[string(base.Bloom.Nonces[t])] {
						return errors.New("Zether Nonce exists")
					}
					nonceMap[string(base.Bloom.Nonces[t])] = true
				}
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
		chainData.ConsecutiveSelfForged,                //atomic copy
	}

	insertedBlocks := []*block_complete.BlockComplete{}

	//remove blocks which are different
	insertedTxs := make(map[string]*transaction.Transaction)

	removedTxsList := make([][]byte, 0)                    //ordered list
	insertedTxsList := make([]*transaction.Transaction, 0) //ordered list
	allTransactionsChanges := make([]*blockchain_types.BlockchainTransactionUpdate, 0)

	var dataStorage *data_storage.DataStorage

	err = func() (err error) {

		chain.mempool.SuspendProcessingCn <- struct{}{}

		err = store.StoreBlockchain.DB.Update(func(writer store_db_interface.StoreDBTransactionInterface) (err error) {

			savedBlock := false

			dataStorage = data_storage.NewDataStorage(writer)

			var accs *accounts.Accounts
			if accs, err = dataStorage.AccsCollection.GetMap(config_coins.NATIVE_ASSET_FULL); err != nil {
				return
			}
			if accs.Count != dataStorage.Regs.Count {
				gui.GUI.Log(fmt.Sprintf("accs != regs %d != %d", accs.Count, dataStorage.Regs.Count))
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

					blockHeightNextStr := strconv.FormatUint(index, 10)
					if err = dataStorage.ReadTransitionalChangesFromStore(blockHeightNextStr); err != nil {
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
				} else {
					newChainData = &BlockchainData{}
					if err = newChainData.loadBlockchainInfo(writer, firstBlockComplete.Block.Height); err != nil {
						return
					}
				}

				if err = dataStorage.CommitChanges(); err != nil {
					return
				}

			}
			if accs.Count != dataStorage.Regs.Count {
				gui.GUI.Log(fmt.Sprintf("accs != regs %d != %d", accs.Count, dataStorage.Regs.Count))
			}

			if blocksComplete[0].Block.Height != newChainData.Height {
				return errors.New("First block hash is not matching")
			}

			if !bytes.Equal(firstBlockComplete.Block.PrevHash, newChainData.Hash) {
				return fmt.Errorf("First block hash is not matching chain hash %d %s %s ", firstBlockComplete.Block.Height, hex.EncodeToString(firstBlockComplete.Bloom.Hash), hex.EncodeToString(newChainData.Hash))
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
					if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(blkComplete.Block.Forger, blkComplete.Block.Height); err != nil {
						return
					}

					if plainAcc == nil {
						return errors.New("Forger Account deson't exist or hasn't delegated stake")
					}

					var stakingAmount uint64
					stakingAmount, err = plainAcc.DelegatedStake.ComputeDelegatedStakeAvailable(newChainData.Height)
					if err != nil {
						return
					}

					if !bytes.Equal(blkComplete.Block.DelegatedStakePublicKey, plainAcc.DelegatedStake.DelegatedStakePublicKey) {
						return errors.New("Block Staking Delegated Public Key is not matching")
					}

					if blkComplete.Block.DelegatedStakeFee != plainAcc.DelegatedStake.DelegatedStakeFee {
						return fmt.Errorf("Block Delegated Stake Fee doesn't match %d %d", blkComplete.Block.DelegatedStakeFee, plainAcc.DelegatedStake.DelegatedStakeFee)
					}

					if blkComplete.Block.StakingAmount != stakingAmount {
						return fmt.Errorf("Block Staking Amount doesn't match %d %d", blkComplete.Block.StakingAmount, stakingAmount)
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

					if err = blkComplete.IncludeBlockComplete(dataStorage); err != nil {
						return errors.New("Error including block into Blockchain: " + err.Error())
					}

					//to detect if the savedBlock was done correctly
					savedBlock = false

					if err = chain.saveBlock(blkComplete, dataStorage); err != nil {
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

					newChainData.saveTotalDifficultyExtra(writer)

					writer.Put("chainHash", newChainData.Hash)

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

				if calledByForging {
					newChainData.ConsecutiveSelfForged += 1
				} else {
					newChainData.ConsecutiveSelfForged = 0
				}

				if err = newChainData.saveBlockchain(writer); err != nil {
					panic("Error saving Blockchain " + err.Error())
				}

				if err = chain.saveBlockchainHashmaps(writer, dataStorage); err != nil {
					panic(err)
				}

				newChainData.AssetsCount = dataStorage.Asts.Count
				newChainData.AccountsCount = dataStorage.Regs.Count + dataStorage.PlainAccs.Count
				chain.ChainData.Store(newChainData)

			} else {
				//only rollback
				if err == nil {
					err = errors.New("Rollback")
				}
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
		UpdateAssets:             multicast.NewMulticastChannel(),
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
	chain.UpdateAssets.CloseAll()
	chain.UpdateRegistrations.CloseAll()
	close(chain.NextBlockCreatedCn)
	close(chain.ForgingSolutionCn)
}
