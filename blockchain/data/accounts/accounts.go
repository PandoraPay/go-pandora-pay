package accounts

import (
	"errors"
	"pandora-pay/blockchain/data/accounts/account"
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

	return data.(*account.Account), nil
}

func NewAccounts(tx store_db_interface.StoreDBTransactionInterface, Token []byte) (accounts *Accounts) {

	hashmap := hash_map.CreateNewHashMap(tx, "accounts", cryptography.PublicKeySize, true)

	accounts = &Accounts{
		HashMap: *hashmap,
		Token:   Token,
	}

	accounts.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		var acc = account.NewAccount(key, accounts.Token)
		if err := acc.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
		return acc, nil
	}
	return
}
