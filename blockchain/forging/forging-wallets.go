package forging

import (
	"bytes"
	"encoding/hex"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/store"
	"sync"
)

type forgingWallets struct {
	addresses    []*forgingWalletAddress
	addressesMap map[string]*forgingWalletAddress

	sync.RWMutex
}

type forgingWalletAddress struct {
	delegatedPrivateKey *addresses.PrivateKey
	delegatedPublicKey  [33]byte

	publicKeyHash [20]byte

	account *account.Account
}

var ForgingW = forgingWallets{
	addressesMap: make(map[string]*forgingWalletAddress),
}

func (w *forgingWallets) AddWallet(delegatedPub [33]byte, delegatedPriv [32]byte, pubKeyHash [20]byte) {

	w.Lock()
	defer w.Unlock()

	//make a clone to be memory safe
	var privateKey [32]byte
	var publicKey [33]byte
	var publicKeyHash [20]byte

	copy(privateKey[:], delegatedPriv[:])
	copy(publicKey[:], delegatedPub[:])
	copy(publicKeyHash[:], pubKeyHash[:])

	private := addresses.PrivateKey{Key: privateKey}

	store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

		var accs *accounts.Accounts
		accs, err = accounts.NewAccounts(tx)

		var acc *account.Account
		if acc, err = accs.GetAccount(publicKeyHash); err != nil {
			return
		}

		address := forgingWalletAddress{
			&private,
			publicKey,
			publicKeyHash,
			acc,
		}
		w.addresses = append(w.addresses, &address)
		w.addressesMap[hex.EncodeToString(pubKeyHash[:])] = &address

		return
	})

}

func (w *forgingWallets) UpdateBalanceChanges(accs *accounts.Accounts) {

	w.Lock()

	for k, v := range accs.HashMap.Virtual {

		if w.addressesMap[k] != nil {

			if v.Committed == "update" {
				w.addressesMap[k].account = new(account.Account)
				_ = w.addressesMap[k].account.Deserialize(v.Data)
			} else if v.Committed == "delete" {
				w.addressesMap[k].account = nil
			}

		}

	}

	w.Unlock()
}

func (w *forgingWallets) RemoveWallet(delegatedPublicKey [33]byte) {

	w.Lock()
	defer w.Unlock()

	for i, address := range w.addresses {
		if bytes.Equal(address.delegatedPublicKey[:], delegatedPublicKey[:]) {
			w.addresses = append(w.addresses[:i], w.addresses[:i+1]...)
			return
		}
	}

}

func (w *forgingWallets) loadBalances() error {

	w.Lock()
	defer w.Unlock()

	return store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

		var accs *accounts.Accounts
		accs, err = accounts.NewAccounts(tx)

		for _, address := range w.addresses {

			var account *account.Account
			if account, err = accs.GetAccount(address.publicKeyHash); err != nil {
				return
			}

			address.account = account
		}

		return
	})

}
