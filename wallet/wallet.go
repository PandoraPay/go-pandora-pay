package wallet

import (
	"pandora-pay/blockchain/forging"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
	wallet_address "pandora-pay/wallet/address"
	"sync"
)

type Wallet struct {
	Encryption     *WalletEncryption               `json:"encryption"`
	Version        Version                         `json:"version"`
	Mnemonic       string                          `json:"mnemonic"`
	Seed           helpers.HexBytes                `json:"seed"` //32 byte
	SeedIndex      uint32                          `json:"seedIndex"`
	Count          int                             `json:"count"`
	CountIndex     int                             `json:"countIndex"`
	Addresses      []*wallet_address.WalletAddress `json:"addresses"`
	addressesMap   map[string]*wallet_address.WalletAddress
	forging        *forging.Forging
	mempool        *mempool.Mempool
	updateAccounts *multicast.MulticastChannel
	loaded         bool
	sync.RWMutex   `json:"-"`
}

func createWallet(forging *forging.Forging, mempool *mempool.Mempool, updateAccounts *multicast.MulticastChannel) (wallet *Wallet) {
	wallet = &Wallet{
		Version:        VERSION_SIMPLE,
		forging:        forging,
		mempool:        mempool,
		Count:          0,
		SeedIndex:      1,
		Addresses:      make([]*wallet_address.WalletAddress, 0),
		addressesMap:   make(map[string]*wallet_address.WalletAddress),
		updateAccounts: updateAccounts,
	}
	wallet.Encryption = createEncryption(wallet)
	return
}

func (wallet *Wallet) clearWallet() {
	wallet.Version = VERSION_SIMPLE
	wallet.Mnemonic = ""
	wallet.Seed = nil
	wallet.SeedIndex = 0
	wallet.Count = 0
	wallet.CountIndex = 0
	wallet.SeedIndex = 1
	wallet.loaded = false
	wallet.Addresses = make([]*wallet_address.WalletAddress, 0)
	wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)
	wallet.Encryption = createEncryption(wallet)
}

func CreateWallet(forging *forging.Forging, mempool *mempool.Mempool) (wallet *Wallet, err error) {

	wallet = createWallet(forging, mempool, nil)

	wallet.initWalletCLI()

	if err = wallet.loadWallet(""); err != nil {
		if err.Error() == "cipher: message authentication failed" {
			err = nil
			return
		}
		if err.Error() != "Wallet doesn't exist" {
			return
		}
		if err = wallet.createEmptyWallet(); err != nil {
			return
		}
	}

	return
}

func (wallet *Wallet) InitializeWallet(updateAccounts *multicast.MulticastChannel) {
	wallet.Lock()
	defer wallet.Unlock()

	wallet.updateAccounts = updateAccounts

	if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
		recovery.SafeGo(wallet.updateAccountsChanges)
	}
}
