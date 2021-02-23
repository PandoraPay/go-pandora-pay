package blockchain

import (
	"pandora-pay/crypto"
	"pandora-pay/gui"
)

type Blockchain struct {
	Hash       crypto.Hash
	Height     int64
	Difficulty uint64

	Sync bool
}

var Chain Blockchain

func BlockchainInit() {

	gui.Info("Blockchain init...")

}
