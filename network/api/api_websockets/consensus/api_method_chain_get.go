package consensus

import (
	"encoding/json"
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

func (consensus *Consensus) ChainGet_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return json.Marshal(consensus.GetUpdateNotification(nil))
}
