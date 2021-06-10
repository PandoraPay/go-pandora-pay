package accounts

import (
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store/hash-map"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type Accounts struct {
	hash_map.HashMap `json:"-"`
}

func NewAccounts(tx store_db_interface.StoreDBTransactionInterface) (accounts *Accounts) {
	accounts = &Accounts{
		HashMap: *hash_map.CreateNewHashMap(tx, "accounts", cryptography.PublicKeyHashHashSize),
	}
	accounts.HashMap.Deserialize = func(data []byte) (helpers.SerializableInterface, error) {
		var acc = &account.Account{}
		err := acc.Deserialize(helpers.NewBufferReader(data))
		return acc, err
	}
	return
}

func (accounts *Accounts) GetAccountEvenEmpty(key []byte, chainHeight uint64) (acc *account.Account, err error) {

	data, err := accounts.Get(string(key))
	if err != nil {
		return
	}

	if data == nil {
		return &account.Account{}, nil
	}

	acc = data.(*account.Account)
	err = acc.RefreshDelegatedStake(chainHeight)
	return
}

func (accounts *Accounts) GetAccount(key []byte, chainHeight uint64) (acc *account.Account, err error) {

	data, err := accounts.Get(string(key))
	if data == nil || err != nil {
		return
	}

	acc = data.(*account.Account)

	if err = acc.RefreshDelegatedStake(chainHeight); err != nil {
		return
	}

	return
}

func (accounts *Accounts) UpdateAccount(key []byte, acc *account.Account) {
	if acc.IsAccountEmpty() {
		accounts.Delete(string(key))
		return
	}
	accounts.Update(string(key), acc)
}
