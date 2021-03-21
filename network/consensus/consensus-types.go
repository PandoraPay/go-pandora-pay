package consensus

import (
	"math/big"
)

type ChainUpdateNotification struct {
	End                uint64
	Hash               []byte
	PrevHash           []byte
	BigTotalDifficulty *big.Int
}

type ChainLastUpdate struct {
	BigTotalDifficulty *big.Int
}
