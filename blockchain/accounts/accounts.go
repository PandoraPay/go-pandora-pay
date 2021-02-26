package accounts

import (
	"errors"
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/account"
)

type VirtualAccount struct {
	account *account.Account
	status  string
}

type Accounts struct {
	Bucket  *bbolt.Bucket
	virtual map[[20]byte]*VirtualAccount
}

func CreateNewAccounts(tx *bbolt.Tx) (accounts *Accounts, err error) {

	if tx == nil {
		err = errors.New("DB Transaction is not set")
		return
	}

	accounts = new(Accounts)
	accounts.virtual = make(map[[20]byte]*VirtualAccount)
	accounts.Bucket = tx.Bucket([]byte("Accounts"))
	return

}

func (accounts *Accounts) GetAccountEvenEmpty(publicKeyHash [20]byte) (acc *account.Account, err error) {
	acc, err = accounts.GetAccount(publicKeyHash)
	if err != nil {
		return
	}
	if acc == nil {
		acc = new(account.Account)
	}
	return
}

func (accounts *Accounts) GetAccount(publicKeyHash [20]byte) (acc *account.Account, err error) {

	exists := accounts.virtual[publicKeyHash]
	if exists != nil {
		acc = exists.account
		return
	}

	data := accounts.Bucket.Get(publicKeyHash[:])
	if data == nil {
		accounts.virtual[publicKeyHash] = &VirtualAccount{
			nil,
			"empty",
		}
		return
	} else {
		acc = new(account.Account)
		if _, err = acc.Deserialize(data); err != nil {
			return nil, err
		}
		accounts.virtual[publicKeyHash] = &VirtualAccount{
			acc,
			"view",
		}
	}

	return
}

func (accounts *Accounts) UpdateAccount(publicKeyHash [20]byte, acc *account.Account) (err error) {

	if acc.IsAccountEmpty() {
		return accounts.DeleteAccount(publicKeyHash)
	} else {
		exists := accounts.virtual[publicKeyHash]
		if exists != nil {
			exists.account = acc
			exists.status = "update"
			return
		} else {
			accounts.virtual[publicKeyHash] = &VirtualAccount{
				acc,
				"update",
			}
		}
	}

	return
}

func (accounts *Accounts) DeleteAccount(publicKeyHash [20]byte) (err error) {

	exists := accounts.virtual[publicKeyHash]
	if exists != nil {
		exists.status = "del"
	} else {
		accounts.virtual[publicKeyHash] = &VirtualAccount{
			nil,
			"del",
		}
	}

	return
}

func (accounts *Accounts) Commit() (err error) {

	for k, v := range accounts.virtual {

		if v.status == "del" {
			if err = accounts.Bucket.Delete(k[:]); err != nil {
				return
			}
			v.status = "empty"
		} else if v.status == "update" {
			data := v.account.Serialize()
			if err = accounts.Bucket.Put(k[:], data); err != nil {
				v.status = "view"
			}
		}

	}

	return
}
