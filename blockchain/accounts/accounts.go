package accounts

import (
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
	"pandora-pay/store"
)

type Accounts struct {
	store.HashMap
}

func NewAccounts(tx *bbolt.Tx) *Accounts {
	return &Accounts{
		HashMap: *store.CreateNewHashMap(tx, "Accounts", 20),
	}
}

func (accounts *Accounts) GetAccountEvenEmpty(key []byte) (acc *account.Account) {

	acc = new(account.Account)

	data := accounts.Get(key)
	if data == nil {
		return
	}

	if err := acc.Deserialize(helpers.NewBufferReader(data)); err != nil {
		panic(err)
	}
	return
}

func (accounts *Accounts) GetAccount(key []byte) *account.Account {

	data := accounts.Get(key)
	if data == nil {
		return nil
	}

	acc := new(account.Account)
	if err := acc.Deserialize(helpers.NewBufferReader(data)); err != nil {
		panic(err)
	}

	return acc
}

func (accounts *Accounts) UpdateAccount(key []byte, acc *account.Account) {
	if acc.IsAccountEmpty() {
		accounts.Delete(key)
		return
	}
	accounts.Update(key, acc.SerializeToBytes())
}
