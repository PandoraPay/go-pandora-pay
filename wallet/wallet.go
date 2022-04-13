package wallet

import (
	"pandora-pay/address_balance_decryptor"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/forging"
	"pandora-pay/config"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/wallet/wallet_address"
	"sync"
)

type Wallet struct {
	Encryption              *WalletEncryption               `json:"encryption" msgpack:"encryption"`
	Version                 Version                         `json:"version" msgpack:"version"`
	Mnemonic                string                          `json:"mnemonic" msgpack:"mnemonic"`
	Seed                    []byte                          `json:"seed" msgpack:"seed"` //32 byte
	SeedIndex               uint32                          `json:"seedIndex" msgpack:"seedIndex"`
	Count                   int                             `json:"count" msgpack:"count"`
	CountImportedIndex      int                             `json:"countIndex" msgpack:"countIndex"`
	Addresses               []*wallet_address.WalletAddress `json:"addresses" msgpack:"addresses"`
	Loaded                  bool                            `json:"loaded" msgpack:"loaded"`
	DelegatesCount          int                             `json:"delegatesCount" msgpack:"delegatesCount"`
	addressesMap            map[string]*wallet_address.WalletAddress
	forging                 *forging.Forging
	mempool                 *mempool.Mempool
	addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor
	updateNewChainUpdate    *multicast.MulticastChannel[*blockchain_types.BlockchainUpdates]
	Lock                    sync.RWMutex `json:"-" msgpack:"-"`
}

func createWallet(forging *forging.Forging, mempool *mempool.Mempool, addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor, updateNewChainUpdate *multicast.MulticastChannel[*blockchain_types.BlockchainUpdates]) (wallet *Wallet) {
	wallet = &Wallet{
		forging:                 forging,
		mempool:                 mempool,
		updateNewChainUpdate:    updateNewChainUpdate,
		addressBalanceDecryptor: addressBalanceDecryptor,
	}
	wallet.clearWallet()
	return
}

//must be locked before
func (wallet *Wallet) clearWallet() {
	wallet.Version = VERSION_SIMPLE_HARDENED
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

func CreateWallet(forging *forging.Forging, mempool *mempool.Mempool, addressBalanceDecryptor *address_balance_decryptor.AddressBalanceDecryptor) (*Wallet, error) {

	wallet := createWallet(forging, mempool, addressBalanceDecryptor, nil)

	if err := wallet.loadWallet("", true); err != nil {
		if err.Error() == "cipher: message authentication failed" {
			return wallet, nil
		}
		if err.Error() != "Wallet doesn't exist" {
			return nil, err
		}
		if err = wallet.CreateEmptyWallet(); err != nil {
			return nil, err
		}
	}

	return wallet, nil
}

func (wallet *Wallet) InitializeWallet(updateNewChainUpdate *multicast.MulticastChannel[*blockchain_types.BlockchainUpdates]) {

	wallet.Lock.Lock()
	wallet.updateNewChainUpdate = updateNewChainUpdate
	wallet.Lock.Unlock()

	if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
		wallet.processRefreshWallets()
	}
}
