package blockchain

import (
	"pandora-pay/block"
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

}
