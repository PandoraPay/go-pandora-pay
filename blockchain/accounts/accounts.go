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

	acc = new(account.Account)

	data := accounts.HashMap.Get(key)
	if data == nil {
		return
	}

	err = acc.Deserialize(data)
	return
}

func (accounts *Accounts) GetAccount(key string) (acc *account.Account, err error) {

	data := accounts.HashMap.Get(key)
	if data == nil {
		return
	}

	acc = new(account.Account)
	err = acc.Deserialize(data)
	return
}

func (accounts *Accounts) UpdateAccount(key string, acc *account.Account) {

	if acc.IsAccountEmpty() {
		accounts.HashMap.Delete(key)
	} else {
		data := acc.Serialize()
		accounts.HashMap.Update(key, data)
	}
}

func (accounts *Accounts) ExistsAccount(key string) bool {
	return accounts.HashMap.Exists(key)
}

func (accounts *Accounts) DeleteAccount(key string) {
	accounts.HashMap.Delete(key)
}

func (accounts *Accounts) Commit() error {
	return accounts.HashMap.Commit()
}
