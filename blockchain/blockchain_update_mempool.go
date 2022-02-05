package blockchain

import (
	"bytes"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/recovery"
)

func (queue *BlockchainUpdatesQueue) processBlockchainUpdateMempool() {
	recovery.SafeGo(func() {

		var err error
		for {

			update := <-queue.updatesMempoolCn

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
					if err = tx.Deserialize(helpers.NewBufferReader(txData)); err != nil {
						return
					}
					if err = queue.txsValidator.MarkAsValidatedTx(tx); err != nil {
						return
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
		}

	})
}
