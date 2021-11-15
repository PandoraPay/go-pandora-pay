package plain_accounts

import (
	"errors"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/cryptography"
	hash_map "pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type PlainAccounts struct {
	*hash_map.HashMap `json:"-"`
}

func (plainAccounts *PlainAccounts) CreatePlainAccount(publicKey []byte) (*plain_account.PlainAccount, error) {

	if len(publicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Key is not a valid public key")
	}

	plainAcc := plain_account.NewPlainAccount(publicKey, 0) //index will be set by update
	if err := plainAccounts.Update(string(publicKey), plainAcc); err != nil {
		return nil, err
	}
	return plainAcc, nil
}

func (plainAccounts *PlainAccounts) GetPlainAccount(key []byte, blockHeight uint64) (*plain_account.PlainAccount, error) {

	data, err := plainAccounts.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	plainAcc := data.(*plain_account.PlainAccount)
	if err = plainAcc.RefreshDelegatedStake(blockHeight); err != nil {
		return nil, err
	}

	return plainAcc, nil
}

func NewPlainAccounts(tx store_db_interface.StoreDBTransactionInterface) (plainAccs *PlainAccounts) {

	hashmap := hash_map.CreateNewHashMap(tx, "plainAccs", cryptography.PublicKeySize, false)

	plainAccs = &PlainAccounts{
		HashMap: hashmap,
	}

	plainAccs.HashMap.CreateObject = func(key []byte, index uint64) (hash_map.HashMapElementSerializableInterface, error) {
		return plain_account.NewPlainAccount(key, index), nil
	}

	return
}
