package consensus

import (
	"pandora-pay/blockchain"
	"pandora-pay/network/websocks/connection"
	"sync/atomic"
)

func (consensus *Consensus) GetChainData() *blockchain.BlockchainData {
	newChainDataPtr := atomic.LoadPointer(&consensus.newChainData)
	return (*blockchain.BlockchainData)(newChainDataPtr)
}

func (consensus *Consensus) chainGet(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	newChainData := consensus.GetChainData()

	conn.Send([]byte("chain"), &ChainUpdateNotification{
		End:                newChainData.Height,
		Hash:               newChainData.Hash,
		PrevHash:           newChainData.PrevHash,
		BigTotalDifficulty: newChainData.BigTotalDifficulty,
	})

	return nil, nil
}

func (consensus *Consensus) broadcast(newChainData *blockchain.BlockchainData) {
	consensus.httpServer.Websockets.Broadcast([]byte("chain"), &ChainUpdateNotification{
		End:                newChainData.Height,
		Hash:               newChainData.Hash,
		PrevHash:           newChainData.PrevHash,
		BigTotalDifficulty: newChainData.BigTotalDifficulty,
	})
}
