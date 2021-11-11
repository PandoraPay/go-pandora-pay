package assets

import (
	"errors"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type Assets struct {
	*hash_map.HashMap `json:"-"`
}

func (assets *Assets) GetAsset(key []byte) (*asset.Asset, error) {

	data, err := assets.HashMap.Get(string(key))
	if data == nil || err != nil {
		return nil, err
	}

	return data.(*asset.Asset), nil
}

func (assets *Assets) CreateAsset(key []byte, ast *asset.Asset) (err error) {

	var exists bool
	if exists, err = assets.Exists(string(key)); err != nil {
		return
	}
	if exists {
		return errors.New("Asset already exists")
	}

	return assets.Update(string(key), ast)
}

func NewAssets(tx store_db_interface.StoreDBTransactionInterface) (assets *Assets) {

	hashMap := hash_map.CreateNewHashMap(tx, "assets", config_coins.ASSET_LENGTH, true)

	assets = &Assets{
		hashMap,
	}

	assets.HashMap.CreateObject = func(key []byte) (helpers.SerializableInterface, error) {
		return asset.NewAsset(key), nil
	}

	return
}
