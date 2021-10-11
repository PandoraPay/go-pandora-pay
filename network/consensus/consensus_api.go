package consensus

import (
	"bytes"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
)

func (consensus *Consensus) chainGet(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	conn.SendJSON([]byte("chain-update"), consensus.getUpdateNotification(nil))
	return nil, nil
}

func (consensus *Consensus) chainUpdate(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	chainUpdateNotification := new(ChainUpdateNotification)
	if err := json.Unmarshal(values, &chainUpdateNotification); err != nil {
		return nil, err
	}

	if len(chainUpdateNotification.Hash) != cryptography.HashSize {
		return nil, errors.New("Chain Update Hash Length is invalid")
	}

	chainLastUpdate := consensus.chain.GetChainData()
	if bytes.Equal(chainLastUpdate.Hash, chainUpdateNotification.Hash) {
		return nil, nil
	}

	forkFound, exists := consensus.forks.hashes.Load(string(chainUpdateNotification.Hash))
	if exists {
		fork := forkFound.(*Fork)
		fork.AddConn(conn, true)
		return nil, nil
	}

	compare := chainLastUpdate.BigTotalDifficulty.Cmp(chainUpdateNotification.BigTotalDifficulty)

	if compare == 0 {
		return nil, nil
	} else if compare < 0 {

		fork := &Fork{
			End:                chainUpdateNotification.End,
			Hash:               chainUpdateNotification.Hash,
			HashStr:            string(chainUpdateNotification.Hash),
			PrevHash:           chainUpdateNotification.PrevHash,
			BigTotalDifficulty: chainUpdateNotification.BigTotalDifficulty,
			Initialized:        false,
			Blocks:             make([]*block_complete.BlockComplete, 0),
			conns:              []*connection.AdvancedConnection{conn},
		}

		consensus.forks.addFork(fork)

	} else {
		//let's notify him tha we have a better chain
		conn.SendJSON([]byte("chain-update"), consensus.getUpdateNotification(nil))
	}

	return nil, nil
}

func (consensus *Consensus) broadcastChain(newChainData *blockchain.BlockchainData) {
	consensus.httpServer.Websockets.BroadcastJSON([]byte("chain-update"), consensus.getUpdateNotification(newChainData), map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true, config.CONSENSUS_TYPE_WALLET: true}, advanced_connection_types.UUID_ALL)
}

func (consensus *Consensus) BroadcastTxs(txs []*transaction.Transaction, justCreated, awaitPropagation bool, exceptSocketUUID advanced_connection_types.UUID) (errs []error) {

	errs = make([]error, len(txs))

	var key, value []byte
	if justCreated {
		key = []byte("mem-pool/new-tx")
	} else {
		key = []byte("mem-pool/new-tx-id")
	}

	for i, tx := range txs {
		if tx != nil {
			if justCreated {
				value = tx.Bloom.Serialized
			} else {
				value = tx.Bloom.Hash
			}

			if awaitPropagation {

				out := consensus.httpServer.Websockets.BroadcastAwaitAnswer(key, value, map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true}, exceptSocketUUID, 2*config.WEBSOCKETS_TIMEOUT)
				for j := range out {
					if out[j] != nil && out[j].Err != nil {
						errs[i] = out[j].Err
						break
					}
				}

			} else {
				consensus.httpServer.Websockets.Broadcast(key, value, map[config.ConsensusType]bool{config.CONSENSUS_TYPE_FULL: true}, exceptSocketUUID)
			}
		}
	}

	return
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
