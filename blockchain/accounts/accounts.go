package accounts

import (
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/account"
)

type Accounts struct {
	Bucket *bbolt.Bucket
}

func CreateNewAccounts(tx *bbolt.Tx) (accounts *Accounts) {

	accounts = new(Accounts)
	accounts.Bucket = tx.Bucket([]byte("Accounts"))
	return

}

func (accounts *Accounts) GetAccount(address [20]byte) (acc *account.Account, err error) {

	data := accounts.Bucket.Get(address[:])
	if data == nil {
		return
	}

	acc = new(account.Account)
	_, err = acc.Deserialize(data)

	return acc, err
}

func (accounts *Accounts) UpdateAccount(account *account.Account) error {

	data := account.Serialize()
	return accounts.Bucket.Put(account.PublicKeyHash[:], data)

}
