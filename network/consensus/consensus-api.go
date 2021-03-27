package consensus

import (
	"encoding/json"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/network/websocks/connection"
	"sync/atomic"
)

func (consensus *Consensus) GetChainData() *blockchain.BlockchainData {
	newChainDataPtr := atomic.LoadPointer(&consensus.newChainData)
	return (*blockchain.BlockchainData)(newChainDataPtr)
}

func (consensus *Consensus) chainGet(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	conn.Send([]byte("chain"), consensus.getUpdateNotification(nil))
	return nil, nil
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
			consensus.forks.Unlock()
		}
	} else {
		//let's notify him tha we have a better chain
		conn.Send([]byte("chain"), consensus.getUpdateNotification(nil))
	}

	return nil, nil
}

func (consensus *Consensus) broadcast(newChainData *blockchain.BlockchainData) {
	consensus.httpServer.Websockets.Broadcast([]byte("chain"), consensus.getUpdateNotification(newChainData))
}

func (consensus *Consensus) getUpdateNotification(newChainData *blockchain.BlockchainData) *ChainUpdateNotification {
	if newChainData == nil {
		newChainData = consensus.GetChainData()
	}
	return &ChainUpdateNotification{
		End:                newChainData.Height,
		Hash:               newChainData.Hash,
		PrevHash:           newChainData.PrevHash,
		BigTotalDifficulty: newChainData.BigTotalDifficulty,
	}
}