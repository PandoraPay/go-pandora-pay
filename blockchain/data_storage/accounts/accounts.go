package accounts

import (
	"errors"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type Accounts struct {
	hash_map.HashMap `json:"-"`
	Asset            []byte
}

func (accounts *Accounts) CreateAccount(publicKey []byte) (*account.Account, error) {

	if len(publicKey) != cryptography.PublicKeySize {
		return nil, errors.New("Key is not a valid public key")
	}

	acc, err := account.NewAccount(publicKey, accounts.Asset)
	if err != nil {
		return nil, err
	}
	accounts.Update(string(publicKey), acc)
	return acc, nil
}

func (accounts *Accounts) GetAccount(key []byte) (*account.Account, error) {
	data, err := accounts.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*account.Account), nil
}

func (accounts *Accounts) GetRandomAccount() (*account.Account, error) {
	data, err := accounts.GetRandom()
	if err != nil {
		return nil, err
	}
	return data.(*account.Account), nil
}

func (accounts *Accounts) saveAssetsCount(key []byte, sign bool) (uint64, error) {

	var count uint64
	var err error

	data := accounts.Tx.Get("accounts:assetsCount:" + string(key))
	if data != nil {
		if count, err = helpers.NewBufferReader(data).ReadUvarint(); err != nil {
			return 0, err
		}
	}

	var countOriginal uint64
	if sign {
		countOriginal = count
		count += 1
	} else {
		count -= 1
		countOriginal = count
	}

	if count > 0 {
		w := helpers.NewBufferWriter()
		w.WriteUvarint(count)
		accounts.Tx.Put("accounts:assetsCount:"+string(key), w.Bytes())
	} else {
		accounts.Tx.Delete("accounts:assetsCount:" + string(key))
	}

	return countOriginal, nil
}

func NewAccounts(tx store_db_interface.StoreDBTransactionInterface, AssetId []byte) (accounts *Accounts, err error) {

	if AssetId == nil || len(AssetId) != cryptography.RipemdSize {
		return nil, errors.New("Asset length is invalid")
	}

	hashmap := hash_map.CreateNewHashMap(tx, "accounts_"+string(AssetId), cryptography.PublicKeySize, true)

	accounts = &Accounts{
		HashMap: *hashmap,
		Asset:   AssetId,
	}

	accounts.HashMap.Deserialize = func(key, data []byte) (helpers.SerializableInterface, error) {
		acc, err := account.NewAccount(key, accounts.Asset)
		if err != nil {
			return nil, err
		}
		if err = acc.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
		return acc, nil
	}

	accounts.HashMap.StoredEvent = func(key []byte, element *hash_map.CommittedMapElement) (err error) {

		if !tx.IsWritable() {
			return
		}

		var count uint64
		if count, err = accounts.saveAssetsCount(key, true); err != nil {
			return
		}

		element.Element.(*account.Account).Index = accounts.HashMap.Count

		tx.Put("accounts:assetByIndex:"+string(key)+":"+strconv.FormatUint(count, 10), element.Element.(*account.Account).Asset)
		return nil
	}

	accounts.HashMap.DeletedEvent = func(key []byte) (err error) {

		if !tx.IsWritable() {
			return
		}

		var count uint64
		if count, err = accounts.saveAssetsCount(key, false); err != nil {
			return
		}

		tx.Delete("accounts:assetByIndex:" + string(key) + ":" + strconv.FormatUint(count, 10))
		return nil
	}

	return
}
