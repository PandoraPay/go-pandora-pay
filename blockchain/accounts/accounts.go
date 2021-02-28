package accounts

import (
	"errors"
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/store"
)

type Accounts struct {
	HashMap *store.HashMap
}

func NewAccounts(tx *bbolt.Tx) (accounts *Accounts, err error) {

	if tx == nil {
		err = errors.New("DB Transaction is not set")
		return
	}

	var hashMap *store.HashMap
	if hashMap, err = store.CreateNewHashMap(tx, "Accounts", 20); err != nil {
		return
	}

	accounts = new(Accounts)
	accounts.HashMap = hashMap
	return
}

func (accounts *Accounts) GetAccountEvenEmpty(key [20]byte) (acc *account.Account, err error) {

	acc = new(account.Account)

	data := accounts.HashMap.Get(key[:])
	if data == nil {
		return
	}

	err = acc.Deserialize(data)
	return
}

func (accounts *Accounts) GetAccount(key [20]byte) (acc *account.Account, err error) {

	data := accounts.HashMap.Get(key[:])
	if data == nil {
		return
	}

	acc = new(account.Account)
	err = acc.Deserialize(data)
	return
}

func (accounts *Accounts) UpdateAccount(key [20]byte, acc *account.Account) {
	if acc.IsAccountEmpty() {
		accounts.HashMap.Delete(key[:])
	} else {
		accounts.HashMap.Update(key[:], acc.Serialize())
	}
}

func (accounts *Accounts) ExistsAccount(key [20]byte) bool {
	return accounts.HashMap.Exists(key[:])
}

func (accounts *Accounts) DeleteAccount(key [20]byte) {
	accounts.HashMap.Delete(key[:])
}

func (accounts *Accounts) Commit() error {
	return accounts.HashMap.Commit()
}
