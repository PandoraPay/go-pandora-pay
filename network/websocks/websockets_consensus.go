package websocks

import (
	"context"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/recovery"
	"time"
)

func (websockets *Websockets) broadcastChain(newChainData *blockchain.BlockchainData, ctx context.Context) {
	websockets.BroadcastJSON([]byte("chain-update"), websockets.ApiWebsockets.Consensus.GetUpdateNotification(newChainData), map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true, config.CONSENSUS_TYPE_WALLET: true}, advanced_connection_types.UUID_ALL, ctx)
}

func (websockets *Websockets) BroadcastTxs(txs []*transaction.Transaction, justCreated, awaitPropagation bool, exceptSocketUUID advanced_connection_types.UUID, ctx context.Context) (errs []error) {

	errs = make([]error, len(txs))

	var key, value []byte
	if justCreated {
		key = []byte("mempool/new-tx")
	} else {
		key = []byte("mempool/new-tx-id")
	}

	if ctx == nil {
		factor := time.Duration(1)
		if awaitPropagation {
			factor = 2
		}

		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), factor*config.WEBSOCKETS_TIMEOUT)
		defer cancel()
	}

	for i, tx := range txs {
		if tx != nil {
			if justCreated {
				value = tx.Bloom.Serialized
			} else {
				value = tx.Bloom.Hash
			}

			if awaitPropagation {
				out := websockets.BroadcastAwaitAnswer(key, value, map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, ctx)
				for j := range out {
					if out[j] != nil && out[j].Err != nil {
						errs[i] = out[j].Err
					}
				}

			} else {
				websockets.Broadcast(key, value, map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, ctx)
			}
		}
	}

	return
}

func (websockets *Websockets) initializeConsensus(chain *blockchain.Blockchain, mempool *mempool.Mempool) {

	recovery.SafeGo(func() {

		updateNewChainUpdateListener := chain.UpdateNewChainDataUpdate.AddListener()
		defer chain.UpdateNewChainDataUpdate.RemoveChannel(updateNewChainUpdateListener)

		var cancelOld context.CancelFunc

		for {
			newChainDataUpdateReceived, ok := <-updateNewChainUpdateListener
			if !ok {
				return
			}

			newChainDataUpdate := newChainDataUpdateReceived.(*blockchain.BlockchainDataUpdate)

			if cancelOld != nil { //let's cancel the previous one
				cancelOld()
			}
			ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_TIMEOUT)
			cancelOld = cancel

			//it is safe to read
			recovery.SafeGo(func() {
				websockets.broadcastChain(newChainDataUpdate.Update, ctx)
			})
		}

	})

	mempool.OnBroadcastNewTransaction = func(txs []*transaction.Transaction, justCreated, awaitPropagation bool, exceptSocketUUID advanced_connection_types.UUID) []error {
		return websockets.BroadcastTxs(txs, justCreated, awaitPropagation, exceptSocketUUID, nil)
	}

}
