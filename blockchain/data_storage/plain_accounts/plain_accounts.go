package plain_accounts

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/cryptography"
	hash_map "pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type PlainAccounts struct {
	*hash_map.HashMap
}

//WARNING: should NOT be used manually without being called from DataStorage
func (plainAccounts *PlainAccounts) CreateNewPlainAccount(publicKey []byte) (*plain_account.PlainAccount, error) {
	plainAcc := plain_account.NewPlainAccount(publicKey, 0) //index will be set by update
	if err := plainAccounts.Create(string(publicKey), plainAcc); err != nil {
		return nil, err
	}
	return plainAcc, nil
}

func (plainAccounts *PlainAccounts) GetPlainAccount(key []byte) (*plain_account.PlainAccount, error) {

	data, err := plainAccounts.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*plain_account.PlainAccount), nil
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
