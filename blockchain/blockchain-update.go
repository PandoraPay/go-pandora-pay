package blockchain

import (
	"encoding/hex"
	"fmt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/recovery"
)

type BlockchainDataUpdate struct {
	Update   *BlockchainData
	SyncTime uint64
}

type BlockchainUpdate struct {
	err              error
	newChainData     *BlockchainData
	accs             *accounts.Accounts
	toks             *tokens.Tokens
	removedTxs       [][]byte
	insertedBlocks   []*block_complete.BlockComplete
	insertedTxHashes [][]byte
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

func (queue *BlockchainUpdatesQueue) hasAnySuccess(updates []*BlockchainUpdate) bool {
	for _, update := range updates {
		if update.err == nil {
			return true
		}
	}

	return false
}

func (queue *BlockchainUpdatesQueue) processUpdate(update *BlockchainUpdate, updates []*BlockchainUpdate) (result bool, err error) {

	if update.err != nil {
		if len(updates) == 0 && !queue.hasAnySuccess(updates) {
			queue.chain.mempool.ContinueWork()
			queue.chain.createNextBlockForForging()
		}
		return
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
		if err = tx.Deserialize(helpers.NewBufferReader(txData)); err != nil {
			return
		}
		if err = tx.BloomExtraNow(true); err != nil {
			return
		}
		if _, err = queue.chain.mempool.AddTxToMemPool(tx, update.newChainData.Height, false); err != nil {
			return
		}
	}

	if !queue.hasAnySuccess(updates) {

		//update work for mem pool
		queue.chain.mempool.UpdateWork(update.newChainData.Hash, update.newChainData.Height)

		//create next block and the workers will be automatically reset
		queue.chain.createNextBlockForForging()
	}

	newSyncTime, result := queue.chain.Sync.addBlocksChanged(uint32(len(update.insertedBlocks)), false)

	if !queue.hasAnySuccess(updates) {

		if result {
			queue.chain.Sync.UpdateSyncMulticast.BroadcastAwait(newSyncTime)
		}

		queue.chain.UpdateNewChain.BroadcastAwait(update.newChainData.Height)

		blockchainDataUpdate := &BlockchainDataUpdate{
			update.newChainData,
			newSyncTime,
		}
		queue.chain.UpdateNewChainDataUpdate.BroadcastAwait(blockchainDataUpdate)

		result = true
	}

	return
}

func (queue *BlockchainUpdatesQueue) processQueue() {
	recovery.SafeGo(func() {

		var updates []*BlockchainUpdate
		for {

			update, ok := <-queue.updatesCn
			if !ok {
				return
			}

			updates = []*BlockchainUpdate{}
			updates = append(updates, update)

			finished := false
			for !finished {
				select {
				case newUpdate, ok := <-queue.updatesCn:
					if !ok {
						return
					}
					updates = append(updates, newUpdate)
				default:
					finished = true
				}
			}

			for len(updates) > 0 {
				update = updates[0]
				updates = updates[1:]

				result, err := queue.processUpdate(update, updates)
				if err != nil {
					gui.GUI.Error("Error processUpdate", err)
				}
				if result {
					break
				}
			}

		}

	})
}
