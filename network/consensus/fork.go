package consensus

import (
	"math/big"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/network/websocks/connection"
	"sync"
)

type Fork struct {
	hashes             [][]byte
	prevHashes         [][]byte
	start              uint64
	end                uint64
	bigTotalDifficulty *big.Int
	ready              bool
	processing         bool
	conns              []*connection.AdvancedConnection
	blocks             []*block_complete.BlockComplete
	sync.RWMutex       `json:"-"`
}

func (fork *Fork) mergeFork(fork2 *Fork) {
	for _, hash := range fork2.hashes {
		fork.hashes = append(fork.hashes, hash)
	}
	for _, prevHash := range fork2.prevHashes {
		fork.prevHashes = append(fork.prevHashes, prevHash)
	}
	fork.end = fork2.end
	fork.bigTotalDifficulty = fork2.bigTotalDifficulty
	for _, conn := range fork2.conns {

		found := false
		for _, conn2 := range fork.conns {
			if conn2 == conn {
				found = true
				break
			}
		}
		if !found {
			fork.conns = append(fork.conns, conn)
		}
	}
}

func (fork *Fork) AddConn(conn *connection.AdvancedConnection) {
	fork.Lock()
	defer fork.Unlock()

	for _, conn2 := range fork.conns {
		if conn2 == conn {
			return
		}
	}

	fork.conns = append(fork.conns, conn)
}
