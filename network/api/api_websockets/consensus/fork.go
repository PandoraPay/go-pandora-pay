package consensus

import (
	"math/big"
	"math/rand"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/helpers/linked_list"
	"pandora-pay/network/websocks/connection"
	"sync"
)

type Fork struct {
	BigTotalDifficulty *big.Int                                               `json:"bigTotalDifficulty" msgpack:"bigTotalDifficulty"`
	Initialized        bool                                                   `json:"initialized" msgpack:"initialized"`
	End                uint64                                                 `json:"end" msgpack:"end"`
	Current            uint64                                                 `json:"current" msgpack:"current"`
	Blocks             *linked_list.LinkedList[*block_complete.BlockComplete] `json:"blocks" msgpack:"blocks"`
	Hash               []byte                                                 `json:"hash" msgpack:"hash"`
	HashStr            string                                                 `json:"hashStr" msgpack:"hashStr"`
	PrevHash           []byte                                                 `json:"prevHash" msgpack:"prevHash"`
	conns              []*connection.AdvancedConnection
	errors             int
	sync.RWMutex       `json:"-" msgpack:"-"`
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
