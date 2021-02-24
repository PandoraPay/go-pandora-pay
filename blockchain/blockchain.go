package blockchain

import (
	"pandora-pay/block"
	"pandora-pay/blockchain/genesis"
	"pandora-pay/crypto"
	"pandora-pay/gui"
	"sync"
)

type Blockchain struct {
	Hash       crypto.Hash
	KernelHash crypto.Hash
	Height     uint64
	Difficulty uint64

	Sync bool

	sync.RWMutex
}

var Chain Blockchain

func (chain *Blockchain) AddBlock(block *block.Block) {

	chain.Lock()

	chain.Unlock()

}

func BlockchainInit() {

	gui.Info("Blockchain init...")

	genesis.GenesisInit()

	Chain.Height = 0
	Chain.Hash = genesis.Genesis.Hash
	Chain.KernelHash = genesis.Genesis.KernelHash
	Chain.Difficulty = genesis.Genesis.Difficulty
	Chain.Sync = false

}
