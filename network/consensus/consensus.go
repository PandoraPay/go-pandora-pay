package consensus

import (
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	node_http "pandora-pay/network/server/node-http"
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

		for {
			newChainDataUpdateReceived, ok := <-updateNewChainUpdateListener
			if !ok {
				return
			}

			newChainDataUpdate := newChainDataUpdateReceived.(*blockchain.BlockchainDataUpdate)

			//it is safe to read
			consensus.broadcastChain(newChainDataUpdate.Update)
		}

	})

	recovery.SafeGo(func() {

		newTxsCn := consensus.mempool.NewTransactionMulticast.AddListener()
		defer consensus.mempool.NewTransactionMulticast.RemoveChannel(newTxsCn)

		for {
			multicastChannelDataReceived, ok := <-newTxsCn
			if !ok {
				return
			}

			multicastChannelData := multicastChannelDataReceived.(*multicast.MulticastChannelData)

			newTxs := multicastChannelData.Data.([]*transaction.Transaction)

			//it is safe to read
			consensus.broadcastTxs(newTxs)

			multicastChannelData.Answer <- struct{}{}
		}

	})

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
