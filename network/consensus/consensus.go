package consensus

import (
	"encoding/json"
	"math/rand"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/mempool"
	node_http "pandora-pay/network/server/node-http"
	"pandora-pay/network/websocks/connection"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Consensus struct {
	httpServer      *node_http.HttpServer
	chain           *blockchain.Blockchain
	chainLastUpdate unsafe.Pointer
	mempool         *mempool.Mempool
	forks           *Forks
}

//must be safe to read
func (consensus *Consensus) updateChain(newChainData *blockchain.BlockchainData) {
	chainLastUpdate := ChainLastUpdate{
		BigTotalDifficulty: newChainData.BigTotalDifficulty,
	}
	atomic.StorePointer(&consensus.chainLastUpdate, unsafe.Pointer(&chainLastUpdate))
}

func (consensus *Consensus) execute() {

	go func() {
		for {
			newChainData, ok := <-consensus.chain.UpdateNewChainChannel
			if ok {
				//it is safe to read
				consensus.updateChain(newChainData)
				consensus.httpServer.Websockets.Broadcast([]byte("chain"), &ChainUpdateNotification{
					End:                newChainData.Height,
					Hash:               newChainData.Hash,
					PrevHash:           newChainData.PrevHash,
					BigTotalDifficulty: newChainData.BigTotalDifficulty,
				})
			}

		}
	}()

	//discover forks
	go func() {
		for {
			var fork *Fork
			consensus.forks.RLock()
			if len(consensus.forks.list) > 0 {
				fork = consensus.forks.list[rand.Intn(len(consensus.forks.list))]
			}
			consensus.forks.RUnlock()

			if fork != nil {
				for i := fork.start; i >= 0; i-- {
					prevHash := fork.prevHashes[0]
					fork2Data, exists := consensus.forks.hashes.LoadOrStore(string(prevHash), fork)
					if exists { //let's merge
						fork2 := fork2Data.(*Fork)
						fork2.RLock()
						if !fork2.processing {
							fork2.mergeFork(fork)
						}
						fork2.RUnlock()
						break
					}
					//exists = fork.prevHash
				}
			}

			time.Sleep(100 * time.Second)
		}
	}()

	//initialize first time
	consensus.updateChain(consensus.chain.GetChainData())
}

func (consensus *Consensus) chainUpdate(conn *connection.AdvancedConnection, values []byte) interface{} {

	chainUpdateNotification := new(ChainUpdateNotification)
	if err := json.Unmarshal(values, &chainUpdateNotification); err != nil {
		return nil
	}

	forkFound, exists := consensus.forks.hashes.Load(string(chainUpdateNotification.Hash))
	if exists {
		fork := forkFound.(*Fork)
		fork.AddConn(conn)
		return nil
	}

	chainLastUpdatePointer := atomic.LoadPointer(&consensus.chainLastUpdate)
	chainLastUpdate := (*ChainLastUpdate)(chainLastUpdatePointer)

	if chainLastUpdate.BigTotalDifficulty.Cmp(chainUpdateNotification.BigTotalDifficulty) < 0 {
		fork := &Fork{
			start:              chainUpdateNotification.End,
			end:                chainUpdateNotification.End,
			hashes:             [][]byte{chainUpdateNotification.Hash},
			prevHashes:         [][]byte{chainUpdateNotification.PrevHash},
			bigTotalDifficulty: chainUpdateNotification.BigTotalDifficulty,
			blocks:             make([]*block_complete.BlockComplete, 0),
			conns:              []*connection.AdvancedConnection{conn},
		}
		_, exists := consensus.forks.hashes.LoadOrStore(string(chainUpdateNotification.Hash), fork)
		if !exists {
			consensus.forks.Lock()
			consensus.forks.list = append(consensus.forks.list, fork)
			consensus.forks.RUnlock()
		}
	}

	return nil
}

func CreateConsensus(httpServer *node_http.HttpServer, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Consensus {

	consensus := &Consensus{
		chain:           chain,
		chainLastUpdate: nil,
		mempool:         mempool,
		httpServer:      httpServer,
		forks: &Forks{
			hashes: sync.Map{},
			list:   make([]*Fork, 0),
		},
	}

	consensus.httpServer.ApiWebsockets.GetMap["chain"] = consensus.chainUpdate

	consensus.execute()

	return consensus
}
