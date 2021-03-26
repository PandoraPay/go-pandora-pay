package consensus

import (
	"bytes"
	"encoding/json"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/mempool"
	api_websockets "pandora-pay/network/api/api-websockets"
	node_http "pandora-pay/network/server/node-http"
	"pandora-pay/network/websocks/connection"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type Consensus struct {
	httpServer   *node_http.HttpServer
	chain        *blockchain.Blockchain
	newChainData unsafe.Pointer
	mempool      *mempool.Mempool
	forks        *Forks
}

func (consensus *Consensus) processFork(fork *Fork) {

	fork.Lock()
	defer fork.Lock()

	if !fork.ready {
		return
	}
	prevHash := fork.prevHash

	for i := fork.start; i >= 0; i-- {

		fork2Data, exists := consensus.forks.hashes.LoadOrStore(string(prevHash), fork)
		if exists { //let's merge
			fork2 := fork2Data.(*Fork)
			if fork2.mergeFork(fork) {
				consensus.forks.removeFork(fork)
				return
			}
		}

		conn := fork.conns[0]
		answer := conn.SendAwaitAnswer([]byte("block-complete"), api_websockets.APIBlockHeight(i-1))
		if answer.Err != nil {
			fork.errors += 1
			if fork.errors > 2 {
				consensus.forks.removeFork(fork)
				return
			}
		} else {
			prevHash := answer.Out

			chainHash, err := consensus.httpServer.ApiWebsockets.ApiStore.LoadBlockHash(i - 1)
			if err == nil {
				if bytes.Equal(prevHash, chainHash) {
					fork.ready = true
					return
				}
			}

			fork.start -= 1
			if fork.errors >= -10 {
				fork.errors -= 1
			}
		}
	}

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
	go func() {
		for {

			fork := consensus.forks.getBestFork()
			if fork != nil {
				consensus.processFork(fork)
			}

			time.Sleep(100 * time.Millisecond)
		}
	}()

	//initialize first time
	newChainData := consensus.chain.GetChainData()
	atomic.StorePointer(&consensus.newChainData, unsafe.Pointer(newChainData)) //newChainData already a pointer
}

func (consensus *Consensus) chainUpdate(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	chainUpdateNotification := new(ChainUpdateNotification)
	if err := json.Unmarshal(values, &chainUpdateNotification); err != nil {
		return nil, err
	}

	forkFound, exists := consensus.forks.hashes.Load(string(chainUpdateNotification.Hash))
	if exists {
		fork := forkFound.(*Fork)
		fork.AddConn(conn)
		return nil, nil
	}

	chainLastUpdate := consensus.GetChainData()

	if chainLastUpdate.BigTotalDifficulty.Cmp(chainUpdateNotification.BigTotalDifficulty) < 0 {
		fork := &Fork{
			start:              chainUpdateNotification.End,
			end:                chainUpdateNotification.End,
			hashes:             [][]byte{chainUpdateNotification.Hash},
			prevHash:           chainUpdateNotification.PrevHash,
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

	return nil, nil
}

func CreateConsensus(httpServer *node_http.HttpServer, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Consensus {

	consensus := &Consensus{
		chain:      chain,
		mempool:    mempool,
		httpServer: httpServer,
		forks: &Forks{
			hashes: sync.Map{},
			list:   make([]*Fork, 0),
		},
	}

	consensus.httpServer.ApiWebsockets.GetMap["chain"] = consensus.chainUpdate
	consensus.httpServer.ApiWebsockets.GetMap["chain-get"] = consensus.chainGet

	consensus.execute()

	return consensus
}
