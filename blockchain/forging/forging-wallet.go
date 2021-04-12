package forging

import (
	"bytes"
	"errors"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/store"
	"sync"
)

type ForgingWallet struct {
	addresses    []*ForgingWalletAddress
	addressesMap map[string]*ForgingWalletAddress
	sync.RWMutex `json:"-"`
}

type ForgingWalletAddress struct {
	delegatedPrivateKey    *addresses.PrivateKey
	delegatedPublicKeyHash []byte //20 byte
	publicKeyHash          []byte //20byte
	account                *account.Account
}

func (w *ForgingWallet) AddWallet(delegatedPriv []byte, pubKeyHash []byte) error {

	if delegatedPriv != nil {
		return nil
	}

	w.Lock()
	defer w.Unlock()

	delegatedPrivateKey := &addresses.PrivateKey{Key: delegatedPriv}

	delegatedPublicKey, err := delegatedPrivateKey.GeneratePublicKey()
	if err != nil {
		return err
	}

	delegatedPublicKeyHash := cryptography.ComputePublicKeyHash(delegatedPublicKey)

	//let's read the balance
	return store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) (err error) {

		accs := accounts.NewAccounts(boltTx)
		acc := accs.GetAccount(pubKeyHash)

		if acc.DelegatedStake == nil || !bytes.Equal(acc.DelegatedStake.DelegatedPublicKeyHash, delegatedPublicKeyHash) {
			return errors.New("Delegated stake is not matching")
		}

		address := w.addressesMap[string(pubKeyHash)]
		if address == nil {
			address = &ForgingWalletAddress{
				delegatedPrivateKey,
				delegatedPublicKeyHash,
				pubKeyHash,
				acc,
			}
			w.addresses = append(w.addresses, address)
			w.addressesMap[string(pubKeyHash)] = address
		} else {
			address.delegatedPrivateKey = delegatedPrivateKey
			address.delegatedPublicKeyHash = delegatedPublicKeyHash
		}

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

func (w *ForgingWallet) RemoveWallet(delegatedPublicKeyHash []byte) { //20 byte

	w.Lock()
	defer w.Unlock()

	for i, address := range w.addresses {
		if bytes.Equal(address.delegatedPublicKeyHash, delegatedPublicKeyHash) {
			w.addresses = append(w.addresses[:i], w.addresses[:i+1]...)
			return
		}
	}

}

func (w *ForgingWallet) loadBalances() error {

	w.Lock()
	defer w.Unlock()

	return store.StoreBlockchain.DB.View(func(boltTx *bolt.Tx) error {

		accs := accounts.NewAccounts(boltTx)

		for _, address := range w.addresses {
			account := accs.GetAccount(address.publicKeyHash)
			address.account = account
		}

		return nil
	})

}
