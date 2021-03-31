package consensus

import (
	"math/big"
	"math/rand"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/network/websocks/connection"
	"sync"
)

type Fork struct {
	downloaded bool

	end     uint64
	current uint64
	blocks  []*block_complete.BlockComplete

	conns []*connection.AdvancedConnection

	hashes             [][]byte
	prevHash           []byte
	bigTotalDifficulty *big.Int
	errors             int
	sync.RWMutex       `json:"-"`
}

//is locked before
func (fork *Fork) getRandomConn() (conn *connection.AdvancedConnection) {

	for len(fork.conns) > 0 {
		index := rand.Intn(len(fork.conns))
		conn = fork.conns[index]
		if conn.IsClosed.IsSet() {
			fork.conns[index] = fork.conns[len(fork.conns)-1]
			fork.conns = fork.conns[:len(fork.conns)-1]
		} else {
			return
		}
	}
	return nil
}

//fork2 must be locked before
func (fork *Fork) mergeFork(fork2 *Fork) bool {

	fork.Lock()
	defer fork.Unlock()

	for _, hash := range fork2.hashes {
		fork.hashes = append(fork.hashes, hash)
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
	return true
}

func (fork *Fork) AddConn(conn *connection.AdvancedConnection, isLocked bool) {

	if !isLocked {
		fork.Lock()
		defer fork.Unlock()
	}

	for _, conn2 := range fork.conns {
		if conn2 == conn {
			return
		}
	}

	fork.conns = append(fork.conns, conn)
}
