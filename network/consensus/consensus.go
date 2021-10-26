package consensus

import (
	"context"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/network/server/node_http"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/recovery"
	"sync"
)

type Consensus struct {
	httpServer *node_http.HttpServer
	chain      *blockchain.Blockchain
	mempool    *mempool.Mempool
	forks      *Forks
}

func (consensus *Consensus) execute() {

	recovery.SafeGo(func() {

		updateNewChainUpdateListener := consensus.chain.UpdateNewChainDataUpdate.AddListener()
		defer consensus.chain.UpdateNewChainDataUpdate.RemoveChannel(updateNewChainUpdateListener)

		var ctx context.Context
		var cancel context.CancelFunc

		for {
			newChainDataUpdateReceived, ok := <-updateNewChainUpdateListener
			if !ok {
				return
			}

			newChainDataUpdate := newChainDataUpdateReceived.(*blockchain.BlockchainDataUpdate)

			if ctx != nil { //let's cancel the previous one
				cancel()
				ctx = nil
			}
			ctx, cancel = context.WithTimeout(context.Background(), config.WEBSOCKETS_TIMEOUT)

			//it is safe to read
			go func() {
				consensus.broadcastChain(newChainDataUpdate.Update, ctx)
			}()
		}

	})

	consensus.mempool.OnBroadcastNewTransaction = func(txs []*transaction.Transaction, justCreated, awaitPropagation bool, exceptSocketUUID advanced_connection_types.UUID) []error {
		return consensus.BroadcastTxs(txs, justCreated, awaitPropagation, exceptSocketUUID, nil)
	}

	//discover forks
	processForksThread := createConsensusProcessForksThread(consensus.forks, consensus.chain, consensus.mempool, consensus.httpServer.ApiStore)
	recovery.SafeGo(processForksThread.execute)

}

func CreateConsensus(httpServer *node_http.HttpServer, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Consensus {

	consensus := &Consensus{
		chain:      chain,
		mempool:    mempool,
		httpServer: httpServer,
		forks: &Forks{
			hashes: &sync.Map{},
		},
	}
	consensus.httpServer.ApiWebsockets.GetMap["chain-update"] = consensus.chainUpdate
	consensus.httpServer.ApiWebsockets.GetMap["chain-get"] = consensus.chainGet

	consensus.execute()

	return consensus
}
