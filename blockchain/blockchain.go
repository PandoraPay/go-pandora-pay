package blockchain

import (
	"errors"
	"math/big"
	"pandora-pay/block"
	"pandora-pay/block/difficulty"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"sync"
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

func (chain *Blockchain) AddBlock(block *block.Block) (result bool, err error) {

	result = false

	chain.Lock()
	defer chain.Unlock() //when the function exists

	if difficulty.CheckKernelHashBig(block.ComputeKernelHash(), Chain.BigDifficulty) != true {
		err = errors.New("KernelHash Difficulty is not met")
		return
	}

	if block.VerifySignature() != true {
		err = errors.New("Forger Signature is invalid!")
		return
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
