package blockchain

import (
	"encoding/base64"
	"fmt"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/recovery"
	"pandora-pay/txs_validator"
)

type BlockchainDataUpdate struct {
	Update        *BlockchainData
	ChainSyncData *blockchain_sync.BlockchainSyncData
}

type BlockchainUpdate struct {
	err                    error
	newChainData           *BlockchainData
	dataStorage            *data_storage.DataStorage
	allTransactionsChanges []*blockchain_types.BlockchainTransactionUpdate
	removedTxHashes        map[string][]byte
	removedTxsList         [][]byte //ordered kept
	insertedTxs            map[string]*transaction.Transaction
	insertedTxsList        []*transaction.Transaction
	insertedBlocks         []*block_complete.BlockComplete
	calledByForging        bool
	exceptSocketUUID       advanced_connection_types.UUID
}

type BlockchainUpdatesQueue struct {
	updatesCn            chan *BlockchainUpdate //buffered
	updatesMempool       *multicast.MulticastChannel[*BlockchainUpdate]
	updatesNotifications *multicast.MulticastChannel[*BlockchainUpdate]
	chain                *Blockchain
	txsValidator         *txs_validator.TxsValidator
}

func createBlockchainUpdatesQueue(txsValidator *txs_validator.TxsValidator) *BlockchainUpdatesQueue {
	return &BlockchainUpdatesQueue{
		make(chan *BlockchainUpdate, 100),
		multicast.NewMulticastChannel[*BlockchainUpdate](),
		multicast.NewMulticastChannel[*BlockchainUpdate](),
		nil,
		txsValidator,
	}
}

func (queue *BlockchainUpdatesQueue) hasCalledByForging(updates []*BlockchainUpdate) bool {
	for _, update := range updates {
		if update.calledByForging {
			return true
		}
	}
	return false
}

func (queue *BlockchainUpdatesQueue) lastSuccess(updates []*BlockchainUpdate) *BlockchainData {
	for i := len(updates) - 1; i >= 0; i-- {
		if updates[i].err == nil {
			return updates[i].newChainData
		}
	}

	return nil
}

func (queue *BlockchainUpdatesQueue) executeUpdate(update *BlockchainUpdate) (err error) {

	gui.GUI.Warning("-------------------------------------------")
	gui.GUI.Warning(fmt.Sprintf("Included blocks %d | TXs: %d | Hash %s", len(update.insertedBlocks), len(update.insertedTxs), base64.StdEncoding.EncodeToString(update.newChainData.Hash)))
	gui.GUI.Warning(update.newChainData.Height, base64.StdEncoding.EncodeToString(update.newChainData.Hash), update.newChainData.Target.Text(10), update.newChainData.BigTotalDifficulty.Text(10))
	gui.GUI.Warning("-------------------------------------------")
	update.newChainData.updateChainInfo()

	queue.chain.UpdateAccounts.Broadcast(update.dataStorage.AccsCollection)
	queue.chain.UpdatePlainAccounts.Broadcast(update.dataStorage.PlainAccs)
	queue.chain.UpdateAssets.Broadcast(update.dataStorage.Asts)
	queue.chain.UpdateRegistrations.Broadcast(update.dataStorage.Regs)

	chainSyncData := queue.chain.Sync.AddBlocksChanged(uint32(len(update.insertedBlocks)), true)

	queue.updatesMempool.Broadcast(update)
	queue.updatesNotifications.Broadcast(update)

	gui.GUI.Log("queue.chain.UpdateNewChain fired")
	queue.chain.UpdateNewChain.Broadcast(update.newChainData.Height)

	queue.chain.UpdateNewChainDataUpdate.Broadcast(&BlockchainDataUpdate{
		update.newChainData,
		chainSyncData,
	})

	return nil
}

func (queue *BlockchainUpdatesQueue) processBlockchainUpdatesQueue() {
	recovery.SafeGo(func() {

		for {

			works := make([]*BlockchainUpdate, 0)
			update, _ := <-queue.updatesCn
			works = append(works, update)

			loop := true
			for loop {
				select {
				case update, _ = <-queue.updatesCn:
					works = append(works, update)
				default:
					loop = false
				}
			}

			lastSuccessUpdate := queue.lastSuccess(works)
			updateForging := lastSuccessUpdate != nil || queue.hasCalledByForging(works)
			for _, update = range works {
				if update.err == nil {
					if err := queue.executeUpdate(update); err != nil {
						gui.GUI.Error("Error processUpdate", err)
					}
				}
			}

			chainSyncData := queue.chain.Sync.GetSyncData()
			if chainSyncData.Started {
				//create next block and the workers will be automatically reset
				queue.chain.createNextBlockForForging(lastSuccessUpdate, updateForging)
			}

		}

	})
}
