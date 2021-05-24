package forging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"sync"
	"sync/atomic"
)

type ForgingWallet struct {
	addresses    []*ForgingWalletAddress
	addressesMap map[string]*ForgingWalletAddress
	sync.RWMutex `json:"-"`

	updates      *atomic.Value //[]*ForgingWalletAddressUpdate
	updatesMutex *sync.Mutex
}

type ForgingWalletAddressUpdate struct {
	delegatedPriv helpers.HexBytes
	pubKeyHash    helpers.HexBytes
}

type ForgingWalletAddress struct {
	delegatedPrivateKey    *addresses.PrivateKey
	delegatedPublicKeyHash helpers.HexBytes //20 byte
	publicKeyHash          helpers.HexBytes //20byte
	account                *account.Account
}

func (w *ForgingWallet) AddWallet(delegatedPriv []byte, pubKeyHash []byte) {

	w.updatesMutex.Lock()
	defer w.updatesMutex.Unlock()

	updates := w.updates.Load().([]*ForgingWalletAddressUpdate)
	updates = append(updates, &ForgingWalletAddressUpdate{
		delegatedPriv,
		pubKeyHash,
	})

	w.updates.Store(updates)

	return
}

func (w *ForgingWallet) RemoveWallet(delegatedPublicKeyHash []byte) { //20 byte
	w.AddWallet(nil, delegatedPublicKeyHash)
}

func (w *ForgingWallet) ProcessUpdates() (err error) {

	w.updatesMutex.Lock()
	updates := w.updates.Load().([]*ForgingWalletAddressUpdate)
	w.updates.Store([]*ForgingWalletAddressUpdate{}) //reset with empty
	w.updatesMutex.Unlock()

	w.Lock()
	defer w.Unlock()

	for _, update := range updates {

		key := string(update.pubKeyHash)

		//let's delete it
		if update.delegatedPriv == nil {

			if w.addressesMap[key] != nil {
				delete(w.addressesMap, key)
				for i, address := range w.addresses {
					if bytes.Equal(address.publicKeyHash, update.pubKeyHash) {
						w.addresses = append(w.addresses[:i], w.addresses[:i+1]...)
						break
					}
				}
			}

		} else {

			delegatedPrivateKey := &addresses.PrivateKey{Key: update.delegatedPriv}

			var delegatedPublicKey []byte
			if delegatedPublicKey, err = delegatedPrivateKey.GeneratePublicKey(); err != nil {
				return err
			}

			delegatedPublicKeyHash := cryptography.ComputePublicKeyHash(delegatedPublicKey)

			//let's read the balance
			store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

				chainHeight, _ := binary.Uvarint(reader.Get([]byte("chainHeight")))

				accs := accounts.NewAccounts(reader)
				var acc *account.Account

				if acc, err = accs.GetAccount(update.pubKeyHash, chainHeight); err != nil {
					return
				}

				if acc == nil {
					return errors.New("Account was not found")
				}

				if acc.DelegatedStake == nil || !bytes.Equal(acc.DelegatedStake.DelegatedPublicKeyHash, delegatedPublicKeyHash) {
					return errors.New("Delegated stake is not matching")
				}

				address := w.addressesMap[key]
				if address == nil {
					address = &ForgingWalletAddress{
						delegatedPrivateKey,
						delegatedPublicKeyHash,
						update.pubKeyHash,
						acc,
					}
					w.addresses = append(w.addresses, address)
					w.addressesMap[key] = address
				} else {
					address.delegatedPrivateKey = delegatedPrivateKey
					address.delegatedPublicKeyHash = delegatedPublicKeyHash
				}

				return
			})

		}

	}

	return
}

func (w *ForgingWallet) UpdateAccountsChanges(accs *accounts.Accounts) (err error) {

	w.Lock()
	defer w.Unlock()

	for k, v := range accs.HashMap.Committed {
		if w.addressesMap[k] != nil {

			if v.Commit == "update" {
				acc := new(account.Account)
				if err = acc.Deserialize(helpers.NewBufferReader(v.Data)); err != nil {
					return
				}
				w.addressesMap[k].account = acc
			} else if v.Commit == "delete" {
				w.addressesMap[k].account = nil
			}

		}
	}

	return
}

func (w *ForgingWallet) loadBalances() error {

	w.Lock()
	defer w.Unlock()

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		accs := accounts.NewAccounts(reader)

		for _, address := range w.addresses {

			chainHeight, _ := binary.Uvarint(reader.Get([]byte("chainHeight")))

			var account *account.Account
			if account, err = accs.GetAccount(address.publicKeyHash, chainHeight); err != nil {
				return
			}

			address.account = account
		}

		return nil
	})

}
