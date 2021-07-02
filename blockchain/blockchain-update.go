package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"pandora-pay/blockchain/accounts"
	blockchain_sync "pandora-pay/blockchain/blockchain-sync"
	blockchain_types "pandora-pay/blockchain/blockchain-types"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/recovery"
)

type BlockchainDataUpdate struct {
	Update        *BlockchainData
	ChainSyncData *blockchain_sync.BlockchainSyncData
}

type BlockchainUpdate struct {
	err                   error
	newChainData          *BlockchainData
	accs                  *accounts.Accounts
	toks                  *tokens.Tokens
	allTransactionChanges []*blockchain_types.BlockchainTransactionUpdate
	removedTxs            [][]byte
	insertedBlocks        []*block_complete.BlockComplete
	insertedTxHashes      [][]byte
	calledByForging       bool
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
	gui.GUI.Warning(fmt.Sprintf("Included blocks %d | TXs: %d | Hash %s", len(update.insertedBlocks), len(update.insertedTxHashes), hex.EncodeToString(update.newChainData.Hash)))
	gui.GUI.Warning(update.newChainData.Height, hex.EncodeToString(update.newChainData.Hash), update.newChainData.Target.Text(10), update.newChainData.BigTotalDifficulty.Text(10))
	gui.GUI.Warning("-------------------------------------------")
	update.newChainData.updateChainInfo()

	queue.chain.UpdateAccounts.BroadcastAwait(update.accs)
	queue.chain.UpdateTokens.BroadcastAwait(update.toks)

	for _, txData := range update.removedTxs {
		tx := &transaction.Transaction{}
		if err := tx.Deserialize(helpers.NewBufferReader(txData)); err != nil {
			return false, err
		}
		if err := tx.BloomExtraVerified(); err != nil {
			return false, err
		}
		if _, err := queue.chain.mempool.AddTxToMemPool(tx, update.newChainData.Height, false, false); err != nil {
			return false, err
		}
		for _, change := range update.allTransactionChanges {
			if bytes.Equal(change.TxHash, tx.Bloom.Hash) {
				change.Tx = tx
			}
		}
	}

	queue.chain.UpdateTransactions.BroadcastAwait(update.allTransactionChanges)

	hasAnySuccess := queue.hasAnySuccess(updates[1:])

	chainSyncData := queue.chain.Sync.AddBlocksChanged(uint32(len(update.insertedBlocks)), hasAnySuccess)

	if !hasAnySuccess {

		//create next block and the workers will be automatically reset
		queue.chain.createNextBlockForForging(update.newChainData, queue.hasCalledByForging(updates))

		queue.chain.UpdateNewChain.BroadcastAwait(update.newChainData.Height)

		queue.chain.UpdateNewChainDataUpdate.BroadcastAwait(&BlockchainDataUpdate{
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
