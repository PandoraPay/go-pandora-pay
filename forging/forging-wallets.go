package forging

import (
	"bytes"
	bolt "go.etcd.io/bbolt"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/account"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/store"
	"sync"
)

type ForgingWallets struct {
	addresses []*ForgingWalletAddress
	sync.RWMutex
}

type ForgingWalletAddress struct {
	delegatedPrivateKey *addresses.PrivateKey
	delegatedPublicKey  [33]byte

	publicKeyHash  [20]byte
	stakeAvailable uint64
}

var ForgingW = ForgingWallets{}

func (w *ForgingWallets) AddWallet(delegatedPub [33]byte, delegatedPriv [32]byte, pubKeyHash [20]byte) {

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

	address := ForgingWalletAddress{
		&private,
		publicKey,
		publicKeyHash,
		0,
	}
	w.addresses = append(w.addresses, &address)

}

func (w *ForgingWallets) RemoveWallet(publicKey [33]byte) {

	w.Lock()
	defer w.Unlock()

	for i, address := range w.addresses {
		if bytes.Equal(address.delegatedPublicKey[:], publicKey[:]) {
			w.addresses = append(w.addresses[:i], w.addresses[:i+1]...)
			return
		}
	}

}

func (w *ForgingWallets) loadBalances() error {

	w.Lock()
	defer w.Unlock()

	return store.StoreBlockchain.DB.View(func(tx *bolt.Tx) (err error) {

		var accs *accounts.Accounts
		accs, err = accounts.CreateNewAccounts(tx)

		for _, address := range w.addresses {

			var account *account.Account
			if account, err = accs.GetAccount(address.publicKeyHash); err != nil {
				return
			}

			if account != nil {
				address.stakeAvailable = account.GetDelegatedStakeAvailable(0)
			}

		}

		return
	})

}
