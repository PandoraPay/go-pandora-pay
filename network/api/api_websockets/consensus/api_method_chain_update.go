package consensus

import (
	"bytes"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/cryptography"
	"pandora-pay/network/websocks/connection"
)

func (consensus *Consensus) ChainUpdate_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

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

	hashStr := string(chainUpdateNotification.Hash)

	forkFound, exists := consensus.forks.hashes.Load(hashStr)
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
			HashStr:            hashStr,
			PrevHash:           chainUpdateNotification.PrevHash,
			BigTotalDifficulty: chainUpdateNotification.BigTotalDifficulty,
			Initialized:        false,
			Blocks:             make([]*block_complete.BlockComplete, 0),
			conns:              []*connection.AdvancedConnection{conn},
		}

		consensus.forks.addFork(fork)

	} else {
		//let's notify him tha we have a better chain
		conn.SendJSON([]byte("chain-update"), consensus.GetUpdateNotification(nil), nil)
	}

	return nil, nil

}
