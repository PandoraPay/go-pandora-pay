package accounts

import (
	"errors"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/helpers/advanced_buffers"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type Accounts struct {
	*hash_map.HashMap[*account.Account]
	Asset []byte
}

// WARNING: should NOT be used manually without being called from DataStorage
func (accounts *Accounts) CreateNewAccount(publicKey []byte) (*account.Account, error) {
	acc, err := account.NewAccount(publicKey, 0, accounts.Asset) //will be set by update
	if err != nil {
		return nil, err
	}
	if err = accounts.Create(string(publicKey), acc); err != nil {
		return nil, err
	}
	return acc, nil
}

func (accounts *Accounts) saveAssetsCount(key []byte, sign bool) (uint64, error) {

	var count uint64
	var err error

	data := accounts.Tx.Get("accounts:assetsCount:" + string(key))
	if data != nil {
		if count, err = advanced_buffers.NewBufferReader(data).ReadUvarint(); err != nil {
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
		w := advanced_buffers.NewBufferWriter()
		w.WriteUvarint(count)
		accounts.Tx.Put("accounts:assetsCount:"+string(key), w.Bytes())
	} else {
		accounts.Tx.Delete("accounts:assetsCount:" + string(key))
	}

	return countOriginal, nil
}

func NewAccounts(tx store_db_interface.StoreDBTransactionInterface, AssetId []byte) (accounts *Accounts, err error) {

	if AssetId == nil || len(AssetId) != cryptography.PublicKeyHashSize {
		return nil, errors.New("Asset length is invalid")
	}

	accounts = &Accounts{
		hash_map.CreateNewHashMap[*account.Account](tx, "accounts_"+string(AssetId), cryptography.PublicKeySize, true),
		AssetId,
	}

	accounts.HashMap.CreateObject = func(key []byte, index uint64) (*account.Account, error) {
		return account.NewAccountClear(key, index, accounts.Asset), nil
	}

	accounts.HashMap.StoredEvent = func(key []byte, committed *hash_map.CommittedMapElement[*account.Account], index uint64) (err error) {

		if !tx.IsWritable() {
			return
		}

		var count uint64
		if count, err = accounts.saveAssetsCount(key, true); err != nil {
			return
		}

		tx.Put("accounts:assetByIndex:"+string(key)+":"+strconv.FormatUint(count, 10), committed.Element.Asset)
		return
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
		return
	}

	return
}
