package consensus

import (
	"math/big"
	"pandora-pay/helpers"
)

type ChainUpdateNotification struct {
	End                uint64           `json:"end" msgpack:"end"`
	Hash               helpers.HexBytes `json:"hash" msgpack:"hash"`
	PrevHash           helpers.HexBytes `json:"prevHash" msgpack:"prevHash"`
	BigTotalDifficulty *big.Int         `json:"bigTotalDifficulty" msgpack:"bigTotalDifficulty"`
}

type ChainLastUpdate struct {
	BigTotalDifficulty *big.Int `json:"bigTotalDifficulty" msgpack:"bigTotalDifficulty"`
}
