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

func CreateNewAccounts(tx *bbolt.Tx) (accounts *Accounts, err error) {

	if tx == nil {
		err = errors.New("DB Transaction is not set")
		return
	}

	var hashMap *store.HashMap
	if hashMap, err = store.CreateNewHashMap(tx, "Accounts"); err != nil {
		return
	}

	accounts = new(Accounts)
	accounts.HashMap = hashMap
	return

}

func (accounts *Accounts) GetAccountEvenEmpty(key string) (acc *account.Account, err error) {

	if acc, err = accounts.GetAccount(key); err != nil {
		return
	}
	if acc == nil {
		acc = new(account.Account)
	}
	return
}

func (accounts *Accounts) GetAccount(key string) (acc *account.Account, err error) {

	var data []byte
	if data, err = accounts.HashMap.Get(key); err != nil || data == nil {
		return
	}

	acc = new(account.Account)
	err = acc.Deserialize(data)
	return
}

func (accounts *Accounts) UpdateAccount(key string, acc *account.Account) (err error) {

	if acc.IsAccountEmpty() {
		return accounts.HashMap.Delete(key)
	} else {
		data := acc.Serialize()
		return accounts.HashMap.Update(key, data)
	}
}

func (accounts *Accounts) DeleteAccount(key string) error {
	return accounts.HashMap.Delete(key)
}

func (accounts *Accounts) Commit() error {
	return accounts.HashMap.Commit()
}
