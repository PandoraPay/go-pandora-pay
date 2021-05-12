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

func (accounts *Accounts) GetAccountEvenEmpty(key []byte, chainHeight uint64) (acc *account.Account, err error) {

	acc = new(account.Account)

	data := accounts.Get(key)
	if data == nil {
		return
	}

	if err = acc.Deserialize(helpers.NewBufferReader(data)); err != nil {
		return
	}

	if err = acc.RefreshDelegatedStake(chainHeight); err != nil {
		return
	}

	return
}

func (accounts *Accounts) GetAccount(key []byte, chainHeight uint64) (acc *account.Account, err error) {

	data := accounts.Get(key)
	if data == nil {
		return
	}

	acc = new(account.Account)
	if err = acc.Deserialize(helpers.NewBufferReader(data)); err != nil {
		return
	}

	if err = acc.RefreshDelegatedStake(chainHeight); err != nil {
		return
	}

	return
}

func (accounts *Accounts) UpdateAccount(key []byte, acc *account.Account) {
	if acc.IsAccountEmpty() {
		accounts.Delete(key)
		return
	}
	accounts.Update(key, acc.SerializeToBytes())
}
