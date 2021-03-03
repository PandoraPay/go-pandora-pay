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

func WalletInit() (wallet *Wallet, err error) {

	wallet = new(Wallet)

	err = wallet.loadWallet()
	if err != nil && err.Error() == "Wallet doesn't exist" {
		err = wallet.createEmptyWallet()
	}
	if err != nil {
		return
	}

	initWalletCLI(wallet)

	gui.Log("Initialized Wallet")
	return
}
