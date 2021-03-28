package consensus

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	node_http "pandora-pay/network/server/node-http"
	"sync"
	"sync/atomic"
	"unsafe"
)

type Consensus struct {
	httpServer   *node_http.HttpServer
	chain        *blockchain.Blockchain
	newChainData unsafe.Pointer
	mempool      *mempool.Mempool
	forks        *Forks
}

func (consensus *Consensus) execute() {

	go func() {
		for {
			newChainData, ok := <-consensus.chain.UpdateNewChainChannel
			if ok {
				//it is safe to read
				atomic.StorePointer(&consensus.newChainData, unsafe.Pointer(newChainData)) //newChainData already a pointer
				consensus.broadcast(newChainData)
			}

		}
	}()

	//discover forks
	processForksThread := createConsensusProcessForksThread(consensus.forks, consensus.chain, consensus.httpServer.ApiWebsockets.ApiStore)
	go processForksThread.execute()

	//initialize first time
	newChainData := consensus.chain.GetChainData()
	atomic.StorePointer(&consensus.newChainData, unsafe.Pointer(newChainData)) //newChainData already a pointer
}

func CreateConsensus(httpServer *node_http.HttpServer, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Consensus {

	consensus := &Consensus{
		chain:      chain,
		mempool:    mempool,
		httpServer: httpServer,
		forks: &Forks{
			hashes:           &sync.Map{},
			forksDownloadMap: &sync.Map{},
			id:               0,
		},
	}

	consensus.httpServer.ApiWebsockets.GetMap["chain"] = consensus.chainUpdate
	consensus.httpServer.ApiWebsockets.GetMap["chain-get"] = consensus.chainGet

	consensus.execute()

	return consensus
}
