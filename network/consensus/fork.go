package consensus

import (
	"math/rand"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/network/websocks/connection"
	"sync"
	"sync/atomic"
)

type Fork struct {
	bigTotalDifficulty *atomic.Value // *big.Int

	downloaded bool

	end     uint64
	current uint64
	blocks  []*block_complete.BlockComplete

	conns []*connection.AdvancedConnection

	hash         []byte
	prevHash     []byte
	errors       int
	sync.RWMutex `json:"-"`
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

func (fork *Fork) AddConn(conn *connection.AdvancedConnection, lock bool) {

	if lock {
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
