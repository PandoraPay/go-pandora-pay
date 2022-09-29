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
	tx   store_db_interface.StoreDBTransactionInterface
	maps map[string]*Accounts
	list []hash_map.HashMapInterface
}

func (this *AccountsCollection) SetTx(tx store_db_interface.StoreDBTransactionInterface) {
	this.tx = tx
}

func (this *AccountsCollection) GetAllMaps() map[string]*Accounts {
	return this.maps
}

func (this *AccountsCollection) GetAllHashmaps() []hash_map.HashMapInterface {
	return this.list
}

func (this *AccountsCollection) GetMap(assetId []byte) (*Accounts, error) {

	if len(assetId) != config_coins.ASSET_LENGTH {
		return nil, errors.New("Asset was not found")
	}

	accs := this.maps[string(assetId)]
	if accs == nil {
		var err error
		if accs, err = NewAccounts(this.tx, assetId); err != nil {
			return nil, err
		}
		this.list = append(this.list, accs.HashMap)
		this.maps[string(assetId)] = accs
	}
	return accs, nil
}

func (this *AccountsCollection) GetMapIfExists(assetId []byte) (*Accounts, error) {
	return this.maps[string(assetId)], nil
}

func (this *AccountsCollection) GetAccountAssetsCount(key []byte) (uint64, error) {

	data := this.tx.Get("accounts:assetsCount:" + string(key))
	if data != nil {
		count, err := helpers.NewBufferReader(data).ReadUvarint()
		if err != nil {
			return 0, err
		}
		return count, nil
	}

	return 0, nil
}

func (this *AccountsCollection) GetAccountAssets(key []byte) ([][]byte, error) {

	count, err := this.GetAccountAssetsCount(key)
	if err != nil {
		return nil, err
	}

	out := make([][]byte, count)

	for i := uint64(0); i < count; i++ {
		assetId := this.tx.Get("accounts:assetByIndex:" + string(key) + ":" + strconv.FormatUint(i, 10))
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
		make([]hash_map.HashMapInterface, 0),
	}
}
