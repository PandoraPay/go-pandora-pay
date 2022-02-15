package consensus

import (
	"math/big"
)

type ChainUpdateNotification struct {
	End                uint64   `json:"end" msgpack:"end"`
	Hash               []byte   `json:"hash" msgpack:"hash"`
	PrevHash           []byte   `json:"prevHash" msgpack:"prevHash"`
	BigTotalDifficulty *big.Int `json:"bigTotalDifficulty" msgpack:"bigTotalDifficulty"`
}

type ChainLastUpdate struct {
	BigTotalDifficulty *big.Int `json:"bigTotalDifficulty" msgpack:"bigTotalDifficulty"`
}
