package accounts

import (
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/store"
)

type Accounts struct {
	HashMap *store.HashMap
}

func NewAccounts(tx *bbolt.Tx) (accounts *Accounts) {

	if tx == nil {
		panic("DB Transaction is not set")
	}

	hashMap := store.CreateNewHashMap(tx, "Accounts", 20)

	accounts = new(Accounts)
	accounts.HashMap = hashMap
	return
}

func (accounts *Accounts) GetAccountEvenEmpty(key [20]byte) (acc *account.Account) {

	acc = new(account.Account)

	data := accounts.HashMap.Get(key[:])
	if data == nil {
		return
	}

	acc.Deserialize(data)
	return
}

func (accounts *Accounts) GetAccount(key [20]byte) *account.Account {

	data := accounts.HashMap.Get(key[:])
	if data == nil {
		return nil
	}

	acc := new(account.Account)
	acc.Deserialize(data)

	return acc
}

func (accounts *Accounts) UpdateAccount(key [20]byte, blockHeight uint64, acc *account.Account) {
	acc.RefreshDelegatedStake(blockHeight)
	if acc.IsAccountEmpty() {
		accounts.HashMap.Delete(key[:])
		return
	}
	accounts.HashMap.Update(key[:], acc.Serialize())
}

func (accounts *Accounts) ExistsAccount(key [20]byte) bool {
	return accounts.HashMap.Exists(key[:])
}

func (accounts *Accounts) DeleteAccount(key [20]byte) {
	accounts.HashMap.Delete(key[:])
}

func (accounts *Accounts) Rollback() {
	accounts.HashMap.Rollback()
}

func (accounts *Accounts) Commit() {
	accounts.HashMap.Commit()
}

func (accounts *Accounts) CommitToStore() {
	accounts.HashMap.CommitToStore()
}
