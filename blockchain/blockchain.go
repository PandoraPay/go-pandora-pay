package blockchain

import "pandora-pay/gui"

type Blockchain struct {
	Height     int64
	Difficulty uint64

	Sync bool
}

func BlockchainInit() {

	gui.Info("Blockchain init...")

}
