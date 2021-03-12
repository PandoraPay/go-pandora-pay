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

func (accounts *Accounts) GetAccountEvenEmpty(key []byte) (acc *account.Account) {

	acc = new(account.Account)

	data := accounts.HashMap.Get(key)
	if data == nil {
		return
	}

	acc.Deserialize(data)
	return
}

func (accounts *Accounts) GetAccount(key []byte) *account.Account {

	data := accounts.HashMap.Get(key)
	if data == nil {
		return nil
	}

	acc := new(account.Account)
	acc.Deserialize(data)

	return acc
}

func (accounts *Accounts) UpdateAccount(key []byte, acc *account.Account) {
	if acc.IsAccountEmpty() {
		accounts.HashMap.Delete(key)
		return
	}
	accounts.HashMap.Update(key, acc.Serialize())
}

func (accounts *Accounts) ExistsAccount(key []byte) bool {
	return accounts.HashMap.Exists(key)
}

func (accounts *Accounts) DeleteAccount(key []byte) {
	accounts.HashMap.Delete(key)
}

func (accounts *Accounts) Rollback() {
	accounts.HashMap.Rollback()
}

func (accounts *Accounts) Commit() {
	accounts.HashMap.Commit()
}

func (accounts *Accounts) WriteToStore() {
	accounts.HashMap.WriteToStore()
}

func (accounts *Accounts) WriteTransitionalChangesToStore(prefix string) {
	accounts.HashMap.WriteTransitionalChangesToStore(prefix)
}
func (accounts *Accounts) ReadTransitionalChangesFromStore(prefix string) {
	accounts.HashMap.ReadTransitionalChangesFromStore(prefix)
}
func (accounts *Accounts) DeleteTransitionalChangesFromStore(prefix string) {
	accounts.HashMap.DeleteTransitionalChangesFromStore(prefix)
}
