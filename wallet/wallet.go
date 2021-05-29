package wallet

import (
	"pandora-pay/blockchain/forging"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	wallet_address "pandora-pay/wallet/address"
	"sync"
)

type Wallet struct {
	Encrypted    EncryptedVersion                         `json:"encrypted"`
	Version      Version                                  `json:"version"`
	Mnemonic     string                                   `json:"mnemonic"`
	Seed         helpers.HexBytes                         `json:"seed"` //32 byte
	SeedIndex    uint32                                   `json:"seedIndex"`
	Count        int                                      `json:"count"`
	CountIndex   int                                      `json:"countIndex"`
	Addresses    []*wallet_address.WalletAddress          `json:"addresses"`
	addressesMap map[string]*wallet_address.WalletAddress `json:"-"`
	forging      *forging.Forging                         `json:"-"`
	mempool      *mempool.Mempool                         `json:"-"`
	sync.RWMutex `json:"-"`
}

func createWallet(forging *forging.Forging, mempool *mempool.Mempool) *Wallet {
	return &Wallet{
		Version:   VERSION_SIMPLE,
		Encrypted: ENCRYPTED_VERSION_PLAIN_TEXT,
		forging:   forging,
		mempool:   mempool,

		Count:     0,
		SeedIndex: 1,

		Addresses:    make([]*wallet_address.WalletAddress, 0),
		addressesMap: make(map[string]*wallet_address.WalletAddress),
	}
}

func WalletInit(forging *forging.Forging, mempool *mempool.Mempool) (wallet *Wallet, err error) {

	wallet = createWallet(forging, mempool)

	if err = wallet.loadWallet(); err != nil {
		if err.Error() != "Wallet doesn't exist" {
			return
		}
		if err = wallet.createEmptyWallet(); err != nil {
			return
		}
	}

	wallet.initWalletCLI()

	gui.GUI.Log("Initialized Wallet")
	return
}
