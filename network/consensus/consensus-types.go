package consensus

import (
	"math/big"
	"pandora-pay/helpers"
)

type ChainUpdateNotification struct {
	End                uint64           `json:"end"`
	Hash               helpers.HexBytes `json:"hash"`
	PrevHash           helpers.HexBytes `json:"prevHash"`
	BigTotalDifficulty *big.Int         `json:"bigTotalDifficulty"`
}

type ChainLastUpdate struct {
	BigTotalDifficulty *big.Int `json:"bigTotalDifficulty"`
}
