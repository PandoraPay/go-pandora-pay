package forging_block_work

import (
	"math/big"
	"pandora-pay/blockchain/blocks/block_complete"
)

type ForgingWork struct {
	BlkComplete   *block_complete.BlockComplete
	BlkSerialized []byte
	BlkTimestmap  uint64
	BlkHeight     uint64
	Target        *big.Int
}
