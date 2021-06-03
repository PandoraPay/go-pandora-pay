package forging_block_work

import (
	"math/big"
	block_complete "pandora-pay/blockchain/block-complete"
)

type ForgingWork struct {
	BlkComplete *block_complete.BlockComplete
	Target      *big.Int
}
