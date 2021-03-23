package wallet

import (
	"pandora-pay/blockchain/forging"
	"pandora-pay/gui"
	"sync"
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

type EncryptedVersion int

const (
	PlainText EncryptedVersion = iota
	Encrypted
)

func (e EncryptedVersion) String() string {
	switch e {
	case PlainText:
		return "PlainText"
	case Encrypted:
		return "Encrypted"
	default:
		return "Unknown EncryptedVersion"
	}
}

type Wallet struct {
	Encrypted EncryptedVersion

	Version   Version
	Mnemonic  string
	Seed      []byte //32 byte
	SeedIndex uint32
	Count     int

	Addresses []*WalletAddress

	Checksum []byte //4 byte

	forging *forging.Forging `json:"-"`

	// forging creates multiple threads and it will read the wallet.Addresses
	sync.RWMutex `json:"-"`
}

func WalletInit(forging *forging.Forging) (wallet *Wallet, err error) {

	wallet = &Wallet{
		forging: forging,
	}

	if err = wallet.loadWallet(); err != nil {
		if err.Error() != "Wallet doesn't exist" {
			return
		}
		if err = wallet.createEmptyWallet(); err != nil {
			return
		}
	}

	initWalletCLI(wallet)

	gui.Log("Initialized Wallet")
	return
}
