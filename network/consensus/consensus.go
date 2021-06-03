package consensus

import (
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/mempool"
	node_http "pandora-pay/network/server/node-http"
	"sync"
	"sync/atomic"
)

type Consensus struct {
	httpServer *node_http.HttpServer
	chain      *blockchain.Blockchain
	mempool    *mempool.Mempool
	forks      *Forks
}

func (consensus *Consensus) execute() {

	go func() {

		updateNewChainUpdateListener := consensus.chain.UpdateNewChainDataUpdate.AddListener()
		for {
			newChainDataUpdateReceived, ok := <-updateNewChainUpdateListener
			if !ok {
				return
			}

			newChainDataUpdate := newChainDataUpdateReceived.(*blockchain.BlockchainDataUpdate)

			//it is safe to read
			consensus.broadcastChain(newChainDataUpdate.Update)
		}

	}()

	go func() {

		newTxCn := consensus.mempool.NewTransactionMulticast.AddListener()
		for {
			newTxReceived, ok := <-newTxCn
			if !ok {
				return
			}

			newTx := newTxReceived.(*transaction.Transaction)

			//it is safe to read
			consensus.broadcastTx(newTx)
		}

	}()

	//discover forks
	processForksThread := createConsensusProcessForksThread(consensus.forks, consensus.chain, consensus.httpServer.ApiStore)
	go processForksThread.execute()

}

func CreateConsensus(httpServer *node_http.HttpServer, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Consensus {

	consensus := &Consensus{
		chain:      chain,
		mempool:    mempool,
		httpServer: httpServer,
		forks: &Forks{
			hashes:    &sync.Map{},
			listMutex: &sync.Mutex{},
			list:      &atomic.Value{},
		},
	}
	consensus.forks.list.Store(make([]*Fork, 0))

	consensus.httpServer.ApiWebsockets.GetMap["chain-update"] = consensus.chainUpdate
	consensus.httpServer.ApiWebsockets.GetMap["chain-get"] = consensus.chainGet

	consensus.execute()

	return consensus
}
