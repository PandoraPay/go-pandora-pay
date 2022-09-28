package websocks

import (
	"context"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/recovery"
	"time"
)

func (websockets *Websockets) broadcastChain(newChainData *blockchain.BlockchainData, ctxDuration time.Duration) {
	websockets.BroadcastJSON([]byte("chain-update"), websockets.ApiWebsockets.Consensus.GetUpdateNotification(newChainData), map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true, config.CONSENSUS_TYPE_WALLET: true}, advanced_connection_types.UUID_ALL, ctxDuration)
}

func (websockets *Websockets) BroadcastTxs(txs []*transaction.Transaction, justCreated, awaitPropagation bool, exceptSocketUUID advanced_connection_types.UUID, ctxParent context.Context) []error {

	errs := make([]error, len(txs))

	for i, tx := range txs {

		select {
		case <-ctxParent.Done():
			return errs
		default:
		}

		var timeout time.Duration //default 0
		if awaitPropagation {
			timeout = time.Duration(3) * config.WEBSOCKETS_TIMEOUT
		}

		if justCreated {

			data := &api_common.APIMempoolNewTxRequest{Tx: tx.Bloom.Serialized}

			if awaitPropagation {
				out := websockets.BroadcastJSONAwaitAnswer([]byte("mempool/new-tx"), data, map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, ctxParent, timeout)
				for _, o := range out {
					if o != nil && o.Err != nil {
						errs[i] = o.Err
					}
				}
			} else {
				websockets.BroadcastJSON([]byte("mempool/new-tx"), data, map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, 0)
			}

		} else {
			if awaitPropagation {
				out := websockets.BroadcastAwaitAnswer([]byte("mempool/new-tx-id"), tx.Bloom.Hash, map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, ctxParent, timeout)
				for _, o := range out {
					if o != nil && o.Err != nil {
						errs[i] = o.Err
					}
				}
			} else {
				websockets.Broadcast([]byte("mempool/new-tx-id"), tx.Bloom.Hash, map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, 0)
			}
		}

	}

	return errs
}

func (websockets *Websockets) initializeConsensus(chain *blockchain.Blockchain, mempool *mempool.Mempool) {

	recovery.SafeGo(func() {

		updateNewChainUpdateListener := chain.UpdateNewChainDataUpdate.AddListener()
		defer chain.UpdateNewChainDataUpdate.RemoveChannel(updateNewChainUpdateListener)

		for {
			newChainDataUpdate, ok := <-updateNewChainUpdateListener
			if !ok {
				return
			}

			//it is safe to read
			recovery.SafeGo(func() {
				websockets.broadcastChain(newChainDataUpdate.Update, 0)
			})
		}

	})

	mempool.OnBroadcastNewTransaction = func(txs []*transaction.Transaction, justCreated, awaitPropagation bool, exceptSocketUUID advanced_connection_types.UUID, ctx context.Context) []error {
		return websockets.BroadcastTxs(txs, justCreated, awaitPropagation, exceptSocketUUID, ctx)
	}

}
