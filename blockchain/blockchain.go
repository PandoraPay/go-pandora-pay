package blockchain

import (
	"errors"
	"math/big"
	"pandora-pay/block"
	"pandora-pay/block/difficulty"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/config"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"sync"
	"time"
)

type Blockchain struct {
	Hash          crypto.Hash
	KernelHash    crypto.Hash
	Height        uint64
	Difficulty    uint64
	BigDifficulty *big.Int //named also as target

	Sync bool

	sync.RWMutex
}

var Chain Blockchain

func (chain *Blockchain) AddBlocks(blocks []*block.Block) (result bool, err error) {

	result = false

	chain.Lock()
	defer chain.Unlock() //when the function exists

	for _, blk := range blocks {

		if difficulty.CheckKernelHashBig(blk.ComputeKernelHash(), Chain.BigDifficulty) != true {
			err = errors.New("KernelHash Difficulty is not met")
			return
		}

		if blk.VerifySignature() != true {
			err = errors.New("Forger Signature is invalid!")
			return
		}

		if blk.BlockHeader.Version != 0 {
			err = errors.New("Invalid Version Version")
			return
		}

		if blk.Timestamp > uint64(time.Now().UTC().Unix())+config.NETWORK_TIMESTAMP_DRIFT_MAX {
			err = errors.New("Timestamp is too much into the future")
			return
		}

		if blk.Height == 0 {

			//verify genesis

		} else {

			//verify block

		}

	}

	result = true
	return

}

func BlockchainInit() {

	gui.Info("Blockchain init...")

	genesis.GenesisInit()

	Chain.Height = 0
	Chain.Hash = genesis.Genesis.Hash
	Chain.KernelHash = genesis.Genesis.KernelHash
	Chain.Difficulty = genesis.Genesis.Difficulty
	Chain.BigDifficulty = difficulty.ConvertDifficultyToBig(Chain.Difficulty)
	Chain.Sync = false

}
