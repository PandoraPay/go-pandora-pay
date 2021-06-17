package wallet

import (
	"pandora-pay/blockchain/forging"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
	wallet_address "pandora-pay/wallet/address"
	"strconv"
	"strings"
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
		forging:        forging,
		mempool:        mempool,
		updateAccounts: updateAccounts,
	}
	wallet.clearWallet()
	return
}

//must be locked before
func (wallet *Wallet) clearWallet() {
	wallet.Version = VERSION_SIMPLE
	wallet.Mnemonic = ""
	wallet.Seed = nil
	wallet.SeedIndex = 0
	wallet.Count = 0
	wallet.CountIndex = 0
	wallet.SeedIndex = 1
	wallet.Addresses = make([]*wallet_address.WalletAddress, 0)
	wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)
	wallet.Encryption = createEncryption(wallet)
	wallet.setLoaded(false)
}

//must be locked before
func (wallet *Wallet) setLoaded(newValue bool) {
	wallet.loaded = newValue
	wallet.initWalletCLI()
}

func CreateWallet(forging *forging.Forging, mempool *mempool.Mempool) (*Wallet, error) {

	wallet := createWallet(forging, mempool, nil)

	if err := wallet.loadWallet("", true); err != nil {
		if err.Error() == "cipher: message authentication failed" {
			return wallet, nil
		}
		if err.Error() != "Wallet doesn't exist" {
			return nil, err
		}
		if err = wallet.createEmptyWallet(); err != nil {
			return nil, err
		}
	}

	return wallet, nil
}

func (wallet *Wallet) ProcessWalletArguments() (err error) {

	if globals.Arguments["--wallet-encrypt"] != nil {

		v := strings.Split(globals.Arguments["--wallet-encrypt"].(string), ",")

		var diff int
		if diff, err = strconv.Atoi(v[1]); err != nil {
			return
		}

		if err = wallet.Encryption.Encrypt(v[0], diff); err != nil {
			return
		}
	}

	if globals.Arguments["--wallet-decrypt"] != nil {
		if err = wallet.loadWallet(globals.Arguments["--wallet-encrypt"].(string), true); err != nil {
			return
		}
	}

	if globals.Arguments["--wallet-remove-encryption"] == true {
		if err = wallet.Encryption.RemoveEncryption(); err != nil {
			return
		}
	}

	return
}

func (wallet *Wallet) InitializeWallet(updateAccounts *multicast.MulticastChannel) {
	wallet.Lock()
	wallet.updateAccounts = updateAccounts
	wallet.Unlock()

	if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
		recovery.SafeGo(wallet.updateAccountsChanges)
	}
}
