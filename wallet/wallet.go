package wallet

import (
	"pandora-pay/gui"
)

type Version int

const (
	VersionSimple Version = 0
)

func (e Version) String() string {
	switch e {
	case VersionSimple:
		return "VersionSimple"
	default:
		return "Unknown Version"
	}
}

func WalletInit() (err error) {

	err = loadWallet()
	if err != nil && err.Error() == "Wallet doesn't exist" {
		err = W.createEmptyWallet()
	}
	if err != nil {
		return
	}

	initWalletCLI()

	gui.Log("Initialized Wallet")
	return
}
