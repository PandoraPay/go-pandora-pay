package plain_accounts

import (
	"errors"
	plain_account "pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	hash_map "pandora-pay/store/hash_map"
	store_db_interface "pandora-pay/store/store_db/store_db_interface"
)

type PlainAccounts struct {
	hash_map.HashMap `json:"-"`
}

func (plainAccounts *PlainAccounts) CreatePlainAccount(publicKey []byte) (*plain_account.PlainAccount, error) {

	if len(publicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Key is not a valid public key")
	}

	plainAcc := plain_account.NewPlainAccount(publicKey)
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
		HashMap: *hashmap,
	}

	plainAccs.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		var plainAcc = plain_account.NewPlainAccount(key)
		if err := plainAcc.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
		return plainAcc, nil
	}
	return
}
