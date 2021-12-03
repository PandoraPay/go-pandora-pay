package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"pandora-pay/blockchain/blockchain_sync"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/recovery"
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
	updates *multicast.MulticastChannel //*BlockchainUpdate //buffered
	chain   *Blockchain
}

func createBlockchainUpdatesQueue() *BlockchainUpdatesQueue {
	return &BlockchainUpdatesQueue{
		updates: multicast.NewMulticastChannel(),
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

func (queue *BlockchainUpdatesQueue) hasAnySuccess(updates []*BlockchainUpdate) bool {
	for _, update := range updates {
		if update.err == nil {
			return true
		}
	}

	return false
}

func (queue *BlockchainUpdatesQueue) processUpdate(update *BlockchainUpdate) error {

	gui.GUI.Warning("-------------------------------------------")
	gui.GUI.Warning(fmt.Sprintf("Included blocks %d | TXs: %d | Hash %s", len(update.insertedBlocks), len(update.insertedTxs), hex.EncodeToString(update.newChainData.Hash)))
	gui.GUI.Warning(update.newChainData.Height, hex.EncodeToString(update.newChainData.Hash), update.newChainData.Target.Text(10), update.newChainData.BigTotalDifficulty.Text(10))
	gui.GUI.Warning("-------------------------------------------")
	update.newChainData.updateChainInfo()

	queue.chain.UpdateAccounts.Broadcast(update.dataStorage.AccsCollection)
	queue.chain.UpdatePlainAccounts.Broadcast(update.dataStorage.PlainAccs)
	queue.chain.UpdateAssets.Broadcast(update.dataStorage.Asts)
	queue.chain.UpdateRegistrations.Broadcast(update.dataStorage.Regs)

	//let's remove the transactions from the mempool
	if len(update.insertedTxsList) > 0 {
		hashes := make([]string, len(update.insertedTxsList))
		for i, tx := range update.insertedTxsList {
			if tx != nil {
				hashes[i] = tx.Bloom.HashStr
			}
		}
		queue.chain.mempool.RemoveInsertedTxsFromBlockchain(hashes)
	}

	//let's add the transactions in the mempool
	if len(update.removedTxsList) > 0 {

		removedTxs := make([]*transaction.Transaction, len(update.removedTxsList))
		for i, txData := range update.removedTxsList {
			tx := &transaction.Transaction{}
			if err := tx.Deserialize(helpers.NewBufferReader(txData)); err != nil {
				return err
			}
			if err := tx.BloomExtraVerified(); err != nil {
				return err
			}
			removedTxs[i] = tx
			for _, change := range update.allTransactionsChanges {
				if bytes.Equal(change.TxHash, tx.Bloom.Hash) {
					change.Tx = tx
				}
			}
		}

		queue.chain.mempool.InsertRemovedTxsFromBlockchain(removedTxs, update.newChainData.Height)
	}

	queue.chain.UpdateTransactions.Broadcast(update.allTransactionsChanges)

	chainSyncData := queue.chain.Sync.AddBlocksChanged(uint32(len(update.insertedBlocks)), true)

	//create next block and the workers will be automatically reset
	queue.chain.createNextBlockForForging(update.newChainData, true)

	gui.GUI.Log("queue.chain.UpdateNewChain fired")
	queue.chain.UpdateNewChain.Broadcast(update.newChainData.Height)

	queue.chain.UpdateNewChainDataUpdate.Broadcast(&BlockchainDataUpdate{
		update.newChainData,
		chainSyncData,
	})

	return nil
}

func (queue *BlockchainUpdatesQueue) processQueue() {
	recovery.SafeGo(func() {

		listener := queue.updates.AddListener()
		defer queue.updates.RemoveChannel(listener)

		for {

			data, ok := <-listener
			if !ok {
				return
			}

			update := data.(*BlockchainUpdate)

			if update.err == nil {
				if err := queue.processUpdate(update); err != nil {
					gui.GUI.Error("Error processUpdate", err)
				}
			}

		}

	})
}
