package consensus

import (
	"encoding/json"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/network/websocks/connection"
)

func (consensus *Consensus) chainGet(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	conn.SendJSON([]byte("chain-update"), consensus.getUpdateNotification(nil))
	return nil, nil
}

func (consensus *Consensus) chainUpdate(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	chainUpdateNotification := new(ChainUpdateNotification)
	if err := json.Unmarshal(values, &chainUpdateNotification); err != nil {
		return nil, err
	}

	forkFound, exists := consensus.forks.hashes.Load(string(chainUpdateNotification.Hash))
	if exists {
		fork := forkFound.(*Fork)
		fork.AddConn(conn, true)
		return
	}

	chainLastUpdate := consensus.chain.GetChainData()

	compare := chainLastUpdate.BigTotalDifficulty.Cmp(chainUpdateNotification.BigTotalDifficulty)

	if compare == 0 {
		return
	} else if compare < 0 {

		fork := &Fork{
			End:                chainUpdateNotification.End,
			Hash:               chainUpdateNotification.Hash,
			PrevHash:           chainUpdateNotification.PrevHash,
			BigTotalDifficulty: chainUpdateNotification.BigTotalDifficulty,
			Initialized:        false,
			Blocks:             make([]*block_complete.BlockComplete, 0),
			conns:              []*connection.AdvancedConnection{conn},
		}

		if _, exists := consensus.forks.hashes.LoadOrStore(string(chainUpdateNotification.Hash), fork); exists {
			return
		} else {
			consensus.forks.listMutex.Lock()
			list := consensus.forks.list.Load().([]*Fork)
			consensus.forks.list.Store(append(list, fork))
			consensus.forks.listMutex.Unlock()
		}

	} else {
		//let's notify him tha we have a better chain
		conn.SendJSON([]byte("chain-update"), consensus.getUpdateNotification(nil))
	}

	return
}

func (consensus *Consensus) broadcast(newChainData *blockchain.BlockchainData) {
	consensus.httpServer.Websockets.BroadcastJSON([]byte("chain-update"), consensus.getUpdateNotification(newChainData), map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true})
}

func (consensus *Consensus) broadcastTx(tx *transaction.Transaction) {
	consensus.httpServer.Websockets.Broadcast([]byte("mem-pool/new-tx-id"), tx.Bloom.Hash, map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true})
}

func (consensus *Consensus) getUpdateNotification(newChainData *blockchain.BlockchainData) *ChainUpdateNotification {
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
