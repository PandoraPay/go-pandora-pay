package wallet

import (
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/blockchain/forging"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/wallet/wallet_address"
	"sync"
)

type Wallet struct {
	Encryption          *WalletEncryption               `json:"encryption"`
	Version             Version                         `json:"version"`
	Mnemonic            string                          `json:"mnemonic"`
	Seed                helpers.HexBytes                `json:"seed"` //32 byte
	SeedIndex           uint32                          `json:"seedIndex"`
	Count               int                             `json:"count"`
	CountImportedIndex  int                             `json:"countIndex"`
	Addresses           []*wallet_address.WalletAddress `json:"addresses"`
	Loaded              bool                            `json:"loaded"`
	DelegatesCount      int                             `json:"delegatesCount"`
	addressesMap        map[string]*wallet_address.WalletAddress
	forging             *forging.Forging
	mempool             *mempool.Mempool
	updateAccounts      *multicast.MulticastChannel[*accounts.AccountsCollection]
	updatePlainAccounts *multicast.MulticastChannel[*plain_accounts.PlainAccounts]
	sync.RWMutex        `json:"-"`
}

func createWallet(forging *forging.Forging, mempool *mempool.Mempool, updateAccounts *multicast.MulticastChannel[*accounts.AccountsCollection], updatePlainAccounts *multicast.MulticastChannel[*plain_accounts.PlainAccounts]) (wallet *Wallet) {
	wallet = &Wallet{
		forging:             forging,
		mempool:             mempool,
		updateAccounts:      updateAccounts,
		updatePlainAccounts: updatePlainAccounts,
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
	wallet.CountImportedIndex = 0
	wallet.Addresses = make([]*wallet_address.WalletAddress, 0)
	wallet.addressesMap = make(map[string]*wallet_address.WalletAddress)
	wallet.Encryption = createEncryption(wallet)
	wallet.setLoaded(false)
}

//must be locked before
func (wallet *Wallet) setLoaded(newValue bool) {
	wallet.Loaded = newValue
	wallet.initWalletCLI()
}

func CreateWallet(forging *forging.Forging, mempool *mempool.Mempool) (*Wallet, error) {

	wallet := createWallet(forging, mempool, nil, nil)

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

func (wallet *Wallet) InitializeWallet(updateAccounts *multicast.MulticastChannel[*accounts.AccountsCollection], updatePlainAccounts *multicast.MulticastChannel[*plain_accounts.PlainAccounts]) {
	wallet.Lock()
	wallet.updateAccounts = updateAccounts
	wallet.updatePlainAccounts = updatePlainAccounts
	wallet.Unlock()

	if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
		wallet.updateAccountsChanges()
	}
}
