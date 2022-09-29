package plain_accounts

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/cryptography"
	hash_map "pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type PlainAccounts struct {
	*hash_map.HashMap[*plain_account.PlainAccount]
}

// WARNING: should NOT be used manually without being called from DataStorage
func (plainAccounts *PlainAccounts) CreateNewPlainAccount(publicKey []byte) (*plain_account.PlainAccount, error) {
	plainAcc := plain_account.NewPlainAccount(publicKey, 0) //index will be set by update
	if err := plainAccounts.Create(string(publicKey), plainAcc); err != nil {
		return nil, err
	}
	return plainAcc, nil
}

func NewPlainAccounts(tx store_db_interface.StoreDBTransactionInterface) (this *PlainAccounts) {

	this = &PlainAccounts{
		hash_map.CreateNewHashMap[*plain_account.PlainAccount](tx, "plainAccs", cryptography.PublicKeySize, false),
	}

	this.HashMap.CreateObject = func(key []byte, index uint64) (*plain_account.PlainAccount, error) {
		return plain_account.NewPlainAccount(key, index), nil
	}

	return
}
