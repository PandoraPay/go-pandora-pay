package accounts

import (
	"errors"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/cryptolib"
	"pandora-pay/helpers"
	"pandora-pay/store/hash-map"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
)

type Accounts struct {
	hash_map.HashMap `json:"-"`
}

func NewAccounts(tx store_db_interface.StoreDBTransactionInterface) (accounts *Accounts) {
	accounts = &Accounts{
		HashMap: *hash_map.CreateNewHashMap(tx, "accounts", cryptography.PublicKeySize),
	}
	accounts.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		var acc = &account.Account{PublicKey: key}
		err := acc.Deserialize(helpers.NewBufferReader(data))
		return acc, err
	}
	return
}

func (accounts *Accounts) CreateAccountValid(key, registration []byte) (*account.Account, error) {
	if len(key) != cryptography.PublicKeySize {
		return nil, errors.New("Key is not a valid public key")
	}
	if cryptolib.VerifySignature([]byte("registration"), registration, key) == false {
		return nil, errors.New("Registration is invalid")
	}

	acc := &account.Account{PublicKey: key}
	if err := accounts.Update(string(key), acc); err != nil {
		return nil, err
	}
	return acc, nil
}

//todo remove
func (accounts *Accounts) GetAccountEvenEmpty(key []byte, chainHeight uint64) (*account.Account, error) {

	data, err := accounts.Get(string(key))
	if err != nil {
		return nil, err
	}

	if data == nil {
		return &account.Account{PublicKey: key}, nil
	}

	acc := data.(*account.Account)
	err = acc.RefreshDelegatedStake(chainHeight)
	return acc, err
}

//todo remove
func (accounts *Accounts) GetAccount(key []byte, chainHeight uint64) (*account.Account, error) {

	data, err := accounts.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	acc := data.(*account.Account)
	if err = acc.RefreshDelegatedStake(chainHeight); err != nil {
		return nil, err
	}

	return acc, nil
}

func (accounts *Accounts) UpdateAccount(key []byte, acc *account.Account) error {
	return accounts.Update(string(key), acc)
}
