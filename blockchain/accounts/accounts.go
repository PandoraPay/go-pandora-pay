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

	data := accounts.HashMap.Get(key)
	if data == nil {
		return
	}

	acc.Deserialize(helpers.NewBufferReader(data))
	return
}

func (accounts *Accounts) GetAccount(key []byte) *account.Account {

	data := accounts.HashMap.Get(key)
	if data == nil {
		return nil
	}

	acc := new(account.Account)
	acc.Deserialize(helpers.NewBufferReader(data))

	return acc
}

func (accounts *Accounts) UpdateAccount(key []byte, acc *account.Account) {
	if acc.IsAccountEmpty() {
		accounts.HashMap.Delete(key)
		return
	}
	accounts.HashMap.Update(key, acc.SerializeToBytes())
}
