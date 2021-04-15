package wallet

import (
	"pandora-pay/blockchain/forging"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	wallet_address "pandora-pay/wallet/address"
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
	Seed      helpers.HexBytes //32 byte
	SeedIndex uint32
	Count     int

	Addresses    []*wallet_address.WalletAddress
	AddressesMap map[string]*wallet_address.WalletAddress

	forging *forging.Forging `json:"-"`
	mempool *mempool.Mempool `json:"-"`

	// forging creates multiple threads and it will read the wallet.Addresses
	sync.RWMutex `json:"-"`
}

func WalletInit(forging *forging.Forging, mempool *mempool.Mempool) (wallet *Wallet, err error) {

	wallet = &Wallet{
		forging: forging,
		mempool: mempool,

		Addresses:    make([]*wallet_address.WalletAddress, 0),
		AddressesMap: make(map[string]*wallet_address.WalletAddress),
	}

	if err = wallet.loadWallet(); err != nil {
		if err.Error() != "Wallet doesn't exist" {
			return
		}
		if err = wallet.createEmptyWallet(); err != nil {
			return
		}
	}

	wallet.initWalletCLI()

	gui.Log("Initialized Wallet")
	return
}
