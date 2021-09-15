package accounts

import (
	"errors"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store/hash-map"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type Accounts struct {
	hash_map.HashMap `json:"-"`
	Token            []byte
}

func (accounts *Accounts) CreateAccount(publicKey []byte) (*account.Account, error) {

	if len(publicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Key is not a valid public key")
	}

	acc := account.NewAccount(publicKey, accounts.Token)
	if err := accounts.Update(string(publicKey), acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (accounts *Accounts) GetAccount(key []byte) (*account.Account, error) {

	data, err := accounts.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	acc := data.(*account.Account)
	return acc, nil
}

func (accounts *Accounts) UpdateAccount(key []byte, acc *account.Account) error {
	return accounts.Update(string(key), acc)
}

func NewAccounts(tx store_db_interface.StoreDBTransactionInterface, Token []byte) (accounts *Accounts, err error) {

	hashmap, err := hash_map.CreateNewHashMap(tx, "accounts", cryptography.PublicKeySize, true)
	if err != nil {
		return nil, err
	}

	accounts = &Accounts{
		HashMap: *hashmap,
		Token:   Token,
	}

	accounts.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		var acc = account.NewAccount(key, accounts.Token)
		err := acc.Deserialize(helpers.NewBufferReader(data))
		return acc, err
	}
	return
}
