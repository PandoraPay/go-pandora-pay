package accounts

import (
	"errors"
	"github.com/jinzhu/copier"
	"go.etcd.io/bbolt"
	"pandora-pay/blockchain/account"
)

type VirtualAccount struct {
	Account   *account.Account
	Status    string
	Committed string
}

type Accounts struct {
	Bucket  *bbolt.Bucket
	Virtual map[[20]byte]*VirtualAccount
}

func CreateNewAccounts(tx *bbolt.Tx) (accounts *Accounts, err error) {

	if tx == nil {
		err = errors.New("DB Transaction is not set")
		return
	}

	accounts = new(Accounts)
	accounts.Virtual = make(map[[20]byte]*VirtualAccount)
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

	exists := accounts.Virtual[publicKeyHash]
	if exists != nil {
		acc2 = new(account.Account)
		copier.Copy(acc2, exists.Account)
		return
	}

	data := accounts.Bucket.Get(publicKeyHash[:])
	if data == nil {
		accounts.Virtual[publicKeyHash] = &VirtualAccount{
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
		accounts.Virtual[publicKeyHash] = &VirtualAccount{
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

		exists := accounts.Virtual[publicKeyHash]
		if exists == nil {
			exists = new(VirtualAccount)
			accounts.Virtual[publicKeyHash] = exists
		}
		exists.Account = acc2
		exists.Status = "update"
	}

	return
}

func (accounts *Accounts) DeleteAccount(publicKeyHash [20]byte) (err error) {

	exists := accounts.Virtual[publicKeyHash]
	if exists == nil {
		exists = new(VirtualAccount)
		accounts.Virtual[publicKeyHash] = exists
	}
	exists.Status = "del"
	exists.Account = nil
	return
}

func (accounts *Accounts) Commit() (err error) {

	for k, v := range accounts.Virtual {

		if v.Status == "del" {
			if err = accounts.Bucket.Delete(k[:]); err != nil {
				return
			}
			v.Status = "empty"
			v.Committed = "del"
			v.Account = nil
		} else if v.Status == "update" {
			data := v.Account.Serialize()
			if err = accounts.Bucket.Put(k[:], data); err != nil {
				return
			}
			v.Committed = "update"
			v.Status = "view"
		}

	}

	return
}
