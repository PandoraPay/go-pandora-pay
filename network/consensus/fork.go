package consensus

import (
	"math/big"
	"math/rand"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/network/websocks/connection"
	"sync"
)

type Fork struct {
	hashes              [][]byte
	prevHash            []byte
	start               uint64
	end                 uint64
	bigTotalDifficulty  *big.Int
	errors              int
	readyForInclusion   bool //ready for including into blockchain
	readyForDownloading bool //ready to downloading
	conns               []*connection.AdvancedConnection
	blocks              []*block_complete.BlockComplete
	sync.RWMutex        `json:"-"`
}

func (fork *Fork) getRandomConn() *connection.AdvancedConnection {
	fork.RLock()
	defer fork.RUnlock()
	if len(fork.conns) > 0 {
		return fork.conns[rand.Intn(len(fork.conns))]
	}
	return nil
}

//fork2 must be locked before
func (fork *Fork) mergeFork(fork2 *Fork) bool {

	if fork2.readyForDownloading {
		return false
	}

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
