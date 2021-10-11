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
	updatesCn chan *BlockchainUpdate //buffered
	chain     *Blockchain
}

func createBlockchainUpdatesQueue() *BlockchainUpdatesQueue {
	return &BlockchainUpdatesQueue{
		updatesCn: make(chan *BlockchainUpdate, 100),
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

func (queue *BlockchainUpdatesQueue) processUpdate(update *BlockchainUpdate, updates []*BlockchainUpdate) (bool, error) {

	if update.err != nil {
		if !queue.hasAnySuccess(updates) {
			queue.chain.createNextBlockForForging(nil, queue.hasCalledByForging(updates))
			return true, nil
		}
		return false, nil
	}

	gui.GUI.Warning("-------------------------------------------")
	gui.GUI.Warning(fmt.Sprintf("Included blocks %d | TXs: %d | Hash %s", len(update.insertedBlocks), len(update.insertedTxs), hex.EncodeToString(update.newChainData.Hash)))
	gui.GUI.Warning(update.newChainData.Height, hex.EncodeToString(update.newChainData.Hash), update.newChainData.Target.Text(10), update.newChainData.BigTotalDifficulty.Text(10))
	gui.GUI.Warning("-------------------------------------------")
	update.newChainData.updateChainInfo()

	queue.chain.UpdateAccounts.Broadcast(update.dataStorage.AccsCollection)
	queue.chain.UpdatePlainAccounts.Broadcast(update.dataStorage.PlainAccs)
	queue.chain.UpdateTokens.Broadcast(update.dataStorage.Toks)
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
				return false, err
			}
			if err := tx.BloomExtraVerified(); err != nil {
				return false, err
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

	hasAnySuccess := queue.hasAnySuccess(updates[1:])

	chainSyncData := queue.chain.Sync.AddBlocksChanged(uint32(len(update.insertedBlocks)), hasAnySuccess)

	if !hasAnySuccess {

		//create next block and the workers will be automatically reset
		queue.chain.createNextBlockForForging(update.newChainData, true)

		gui.GUI.Log("queue.chain.UpdateNewChain fired")
		queue.chain.UpdateNewChain.Broadcast(update.newChainData.Height)

		queue.chain.UpdateNewChainDataUpdate.Broadcast(&BlockchainDataUpdate{
			update.newChainData,
			chainSyncData,
		})

		return true, nil
	}

	return false, nil
}

func (queue *BlockchainUpdatesQueue) processQueue() {
	recovery.SafeGo(func() {

		var updates []*BlockchainUpdate
		for {

			updates = []*BlockchainUpdate{}
			exitCn := make(chan struct{})

			finished := false
			for !finished {
				select {
				case newUpdate, ok := <-queue.updatesCn:
					if !ok {
						return
					}
					updates = append(updates, newUpdate)
					if len(updates) == 1 {
						close(exitCn)
					}
				case <-exitCn:
					finished = true
				}
			}

			for len(updates) > 0 {

				result, err := queue.processUpdate(updates[0], updates)

				if err != nil {
					gui.GUI.Error("Error processUpdate", err)
				}
				if result {
					break
				}

				updates = updates[1:]

			}

		}

	})
}
