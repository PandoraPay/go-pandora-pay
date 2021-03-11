package forging

import (
	"bytes"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/store"
	"sync"
)

type ForgingWallet struct {
	addresses    []*ForgingWalletAddress
	addressesMap map[[20]byte]*ForgingWalletAddress

	sync.RWMutex
}

type ForgingWalletAddress struct {
	delegatedPrivateKey *addresses.PrivateKey
	delegatedPublicKey  [33]byte

	publicKeyHash [20]byte

	account *account.Account
}

type ForgingWalletAddressRequired struct {
	publicKeyHash [20]byte
	wallet        *ForgingWalletAddress
	stakingAmount uint64
}

func (w *ForgingWallet) AddWallet(delegatedPub [33]byte, delegatedPriv [32]byte, pubKeyHash [20]byte) {

	w.Lock()
	defer w.Unlock()

	private := addresses.PrivateKey{Key: delegatedPriv}

	store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

		accs := accounts.NewAccounts(tx)

		acc := accs.GetAccount(&pubKeyHash)

		address := ForgingWalletAddress{
			&private,
			delegatedPub,
			pubKeyHash,
			acc,
		}
		w.addresses = append(w.addresses, &address)
		w.addressesMap[pubKeyHash] = &address

		return
	})

}

func (w *ForgingWallet) UpdateBalanceChanges(accs *accounts.Accounts) {

	w.Lock()
	defer w.Unlock()

	for k, v := range accs.HashMap.Committed {

		if w.addressesMap[k] != nil {

			if v.Commit == "update" {
				w.addressesMap[k].account = new(account.Account)
				w.addressesMap[k].account.Deserialize(v.Data)
			} else if v.Commit == "delete" {
				w.addressesMap[k].account = nil
			}

		}

	}

}

func (w *ForgingWallet) RemoveWallet(delegatedPublicKey [33]byte) {

	w.Lock()
	defer w.Unlock()

	for i, address := range w.addresses {
		if bytes.Equal(address.delegatedPublicKey[:], delegatedPublicKey[:]) {
			w.addresses = append(w.addresses[:i], w.addresses[:i+1]...)
			return
		}
	}

}

func (w *ForgingWallet) loadBalances() error {

	w.Lock()
	defer w.Unlock()

	return store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

		accs := accounts.NewAccounts(tx)

		for _, address := range w.addresses {

			account := accs.GetAccount(&address.publicKeyHash)
			address.account = account

		}

		return
	})

}
