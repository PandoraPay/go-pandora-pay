package accounts

import (
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/account"
)

type Accounts struct {
	Bucket    *bbolt.Bucket
	ReadyOnly bool
}

func CreateNewAccounts(tx *bbolt.Tx, ReadyOnly bool) (accounts *Accounts) {

	accounts = new(Accounts)
	accounts.Bucket = tx.Bucket([]byte("Accounts"))
	accounts.ReadyOnly = ReadyOnly
	return

}

func (accounts *Accounts) GetAccount(publicKeyHash [20]byte, createEmptyIfNotFound bool) (acc *account.Account, err error) {

	data := accounts.Bucket.Get(publicKeyHash[:])
	if data == nil {
		if createEmptyIfNotFound {
			acc = new(account.Account)
		}
		return
	}

	acc = new(account.Account)
	_, err = acc.Deserialize(data)

	return
}

func (accounts *Accounts) UpdateAccount(publicKeyHash [20]byte, acc *account.Account) error {

	if acc.IsAccountEmpty() {
		return accounts.Bucket.Delete(publicKeyHash[:])
	} else {
		data := acc.Serialize()
		return accounts.Bucket.Put(publicKeyHash[:], data)
	}

}
