package consensus

import (
	"pandora-pay/blockchain"
	"pandora-pay/network/websocks/connection"
)

func (consensus *Consensus) GetUpdateNotification(newChainData *blockchain.BlockchainData) *ChainUpdateNotification {

	if newChainData == nil {
		newChainData = consensus.chain.GetChainData()
	}

	return &ChainUpdateNotification{
		End:                newChainData.Height,
		Hash:               newChainData.Hash,
		PrevHash:           newChainData.PrevHash,
		BigTotalDifficulty: newChainData.BigTotalDifficulty,
	}
}

func (consensus *Consensus) GetChain(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	return consensus.GetUpdateNotification(nil), nil
}
