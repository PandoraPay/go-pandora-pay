package accounts

import (
	"errors"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type AccountsCollection struct {
	tx       store_db_interface.StoreDBTransactionInterface
	accsMap  map[string]*Accounts
	listMaps []*hash_map.HashMap
}

func (collection *AccountsCollection) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	collection.tx = tx
}

func (collection *AccountsCollection) GetAllMaps() map[string]*Accounts {
	return collection.accsMap
}

func (collection *AccountsCollection) GetAllHashmaps() []*hash_map.HashMap {
	return collection.listMaps
}

func (collection *AccountsCollection) GetMap(assetId []byte) (*Accounts, error) {

	if len(assetId) != config_coins.ASSET_LENGTH {
		return nil, errors.New("Asset was not found")
	}

	accs := collection.accsMap[string(assetId)]
	if accs == nil {
		var err error
		if accs, err = NewAccounts(collection.tx, assetId); err != nil {
			return nil, err
		}
		collection.listMaps = append(collection.listMaps, accs.HashMap)
		collection.accsMap[string(assetId)] = accs
	}
	return accs, nil
}

func (collection *AccountsCollection) GetAccountAssetsCount(key []byte) (uint64, error) {

	data := collection.tx.Get("accounts:assetsCount:" + string(key))
	if data != nil {
		count, err := helpers.NewBufferReader(data).ReadUvarint()
		if err != nil {
			return 0, err
		}
		return count, nil
	}

	return 0, nil
}

func (collection *AccountsCollection) GetAccountAssets(key []byte) ([][]byte, error) {

	count, err := collection.GetAccountAssetsCount(key)
	if err != nil {
		return nil, err
	}

	out := make([][]byte, count)

	for i := uint64(0); i < count; i++ {
		assetId := collection.tx.Get("accounts:assetByIndex:" + string(key) + ":" + strconv.FormatUint(i, 10))
		if assetId == nil {
			return nil, errors.New("Error reading AssetId")
		}
		out[i] = assetId
	}

	return out, nil
}

func NewAccountsCollection(tx store_db_interface.StoreDBTransactionInterface) *AccountsCollection {
	return &AccountsCollection{
		tx,
		make(map[string]*Accounts),
		make([]*hash_map.HashMap, 0),
	}
}
