package consensus

import (
	"encoding/json"
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	node_http "pandora-pay/network/server/node-http"
)

type Consensus struct {
	httpServer *node_http.HttpServer
	chain      *blockchain.Blockchain
	mempool    *mempool.Mempool
}

func (consensus *Consensus) execute() {

	go func() {
		for {
			newchain, ok := <-consensus.chain.UpdateNewChainChannel
			if ok {
				//it is safe to read
				consensus.httpServer.Websockets.Broadcast([]byte("chain"), &ChainUpdateNotification{
					End:                newchain.Height,
					Hash:               newchain.Hash,
					BigTotalDifficulty: newchain.BigTotalDifficulty,
				})
			}

		}
	}()

}

func (consensus *Consensus) chainUpdate(values []byte) interface{} {

	chainUpdateNotification := new(ChainUpdateNotification)
	if err := json.Unmarshal(values, &chainUpdateNotification); err != nil {
		return nil
	}

	return nil
}

func CreateConsensus(httpServer *node_http.HttpServer, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Consensus {

	consensus := &Consensus{
		chain:      chain,
		mempool:    mempool,
		httpServer: httpServer,
	}

	consensus.httpServer.ApiWebsockets.GetMap["chain"] = consensus.chainUpdate

	consensus.execute()

	return consensus
}
