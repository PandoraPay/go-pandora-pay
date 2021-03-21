package consensus

import "math/big"

type ChainUpdateNotification struct {
	End                uint64
	Hash               []byte
	BigTotalDifficulty *big.Int
}
