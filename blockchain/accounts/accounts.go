package accounts

import (
	"errors"
	"github.com/jinzhu/copier"
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/account"
)

type VirtualAccount struct {
	account   *account.Account
	status    string
	committed string
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
	if acc, err = accounts.GetAccount(publicKeyHash); err != nil {
		return
	}
	if acc == nil {
		acc = new(account.Account)
	}
	return
}

func (accounts *Accounts) GetAccount(publicKeyHash [20]byte) (acc2 *account.Account, err error) {

	exists := accounts.virtual[publicKeyHash]
	if exists != nil {
		acc2 = new(account.Account)
		copier.Copy(acc2, exists.account)
		return
	}

	data := accounts.Bucket.Get(publicKeyHash[:])
	if data == nil {
		accounts.virtual[publicKeyHash] = &VirtualAccount{
			nil,
			"empty",
			"",
		}
		return
	} else {
		acc := new(account.Account)
		if _, err = acc.Deserialize(data); err != nil {
			return nil, err
		}
		accounts.virtual[publicKeyHash] = &VirtualAccount{
			acc,
			"view",
			"",
		}
		acc2 = new(account.Account)
		copier.Copy(&acc2, acc)
		return
	}

}

func (accounts *Accounts) UpdateAccount(publicKeyHash [20]byte, acc *account.Account) (err error) {

	if acc.IsAccountEmpty() {
		return accounts.DeleteAccount(publicKeyHash)
	} else {

		var acc2 = new(account.Account)
		copier.Copy(acc2, acc)

		exists := accounts.virtual[publicKeyHash]
		if exists == nil {
			exists = new(VirtualAccount)
			accounts.virtual[publicKeyHash] = exists
		}
		exists.account = acc2
		exists.status = "update"
	}

	return
}

func (accounts *Accounts) DeleteAccount(publicKeyHash [20]byte) (err error) {

	exists := accounts.virtual[publicKeyHash]
	if exists == nil {
		exists = new(VirtualAccount)
		accounts.virtual[publicKeyHash] = exists
	}
	exists.status = "del"
	exists.account = nil
	return
}

func (accounts *Accounts) Commit() (err error) {

	for k, v := range accounts.virtual {

		if v.status == "del" {
			if err = accounts.Bucket.Delete(k[:]); err != nil {
				return
			}
			v.status = "empty"
			v.committed = "del"
			v.account = nil
		} else if v.status == "update" {
			data := v.account.Serialize()
			if err = accounts.Bucket.Put(k[:], data); err != nil {
				return
			}
			v.committed = "update"
			v.status = "view"
		}

	}

	return
}
