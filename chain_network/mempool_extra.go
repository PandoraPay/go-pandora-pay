package chain_network

import (
	"context"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers/recovery"
	"pandora-pay/mempool"
	"pandora-pay/network/api_implementation/api_common"
	"pandora-pay/network/network_config"
	"pandora-pay/network/server/node_http"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"time"
)

func broadcastChain(newChainData *blockchain.BlockchainData, ctxDuration time.Duration) {
	websocks.Websockets.BroadcastJSON([]byte("chain-update"), node_http.HttpServer.ApiWebsockets.Consensus.GetUpdateNotification(newChainData), map[config.NodeConsensusType]bool{config.NODE_CONSENSUS_TYPE_FULL: true, config.NODE_CONSENSUS_TYPE_APP: true}, advanced_connection_types.UUID_ALL, ctxDuration)
}

func BroadcastTxs(txs []*transaction.Transaction, justCreated, awaitPropagation bool, exceptSocketUUID advanced_connection_types.UUID, ctxParent context.Context) []error {

	errs := make([]error, len(txs))

	for i, tx := range txs {

		select {
		case <-ctxParent.Done():
			return errs
		default:
		}

		var timeout time.Duration //default 0
		if awaitPropagation {
			timeout = time.Duration(3) * network_config.WEBSOCKETS_TIMEOUT
		}

		if justCreated {

			data := &api_common.APIMempoolNewTxRequest{Tx: tx.Bloom.Serialized}

			if awaitPropagation {
				out := websocks.Websockets.BroadcastJSONAwaitAnswer([]byte("mempool/new-tx"), data, map[config.NodeConsensusType]bool{config.NODE_CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, ctxParent, timeout)
				for _, o := range out {
					if o != nil && o.Err != nil {
						errs[i] = o.Err
					}
				}
			} else {
				websocks.Websockets.BroadcastJSON([]byte("mempool/new-tx"), data, map[config.NodeConsensusType]bool{config.NODE_CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, 0)
			}

		} else {
			if awaitPropagation {
				out := websocks.Websockets.BroadcastAwaitAnswer([]byte("mempool/new-tx-id"), tx.Bloom.Hash, map[config.NodeConsensusType]bool{config.NODE_CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, ctxParent, timeout)
				for _, o := range out {
					if o != nil && o.Err != nil {
						errs[i] = o.Err
					}
				}
			} else {
				websocks.Websockets.Broadcast([]byte("mempool/new-tx-id"), tx.Bloom.Hash, map[config.NodeConsensusType]bool{config.NODE_CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, 0)
			}
		}

	}

	return errs
}

func initializeConsensus(chain *blockchain.Blockchain, mempool *mempool.Mempool) {

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
				broadcastChain(newChainDataUpdate.Update, 0)
			})
		}

	})

	mempool.OnBroadcastNewTransaction = func(txs []*transaction.Transaction, justCreated, awaitPropagation bool, exceptSocketUUID advanced_connection_types.UUID, ctx context.Context) []error {
		return BroadcastTxs(txs, justCreated, awaitPropagation, exceptSocketUUID, ctx)
	}

}
