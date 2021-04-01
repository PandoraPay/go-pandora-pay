package consensus

import (
	"pandora-pay/blockchain"
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

		updateNewChainCn := consensus.chain.UpdateNewChainMulticast.AddListener()
		for {
			newChainDataReceived, ok := <-updateNewChainCn
			if !ok {
				return
			}
			newChainData := newChainDataReceived.(*blockchain.BlockchainData)

			//it is safe to read
			consensus.broadcast(newChainData)
		}
	}()

	//discover forks
	processForksThread := createConsensusProcessForksThread(consensus.forks, consensus.chain, consensus.httpServer.ApiWebsockets.ApiStore)
	go processForksThread.execute()

}

func CreateConsensus(httpServer *node_http.HttpServer, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Consensus {

	consensus := &Consensus{
		chain:      chain,
		mempool:    mempool,
		httpServer: httpServer,
		forks: &Forks{
			hashes:    &sync.Map{},
			listMutex: sync.Mutex{},
			list:      atomic.Value{},
		},
	}
	consensus.forks.list.Store(make([]*Fork, 0))

	consensus.httpServer.ApiWebsockets.GetMap["chain"] = consensus.chainUpdate
	consensus.httpServer.ApiWebsockets.GetMap["chain-get"] = consensus.chainGet

	consensus.execute()

	return consensus
}
