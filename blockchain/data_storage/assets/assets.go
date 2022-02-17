package assets

import (
	"encoding/hex"
	"errors"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/config/config_coins"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type Assets struct {
	*hash_map.HashMap
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

	assets.HashMap.CreateObject = func(key []byte, index uint64) (hash_map.HashMapElementSerializableInterface, error) {
		return asset.NewAsset(key, index), nil
	}

	assets.HashMap.StoredEvent = func(key []byte, element *hash_map.CommittedMapElement) (err error) {
		if !tx.IsWritable() {
			return
		}

		asset := element.Element.(*asset.Asset)

		tokenIdentification := asset.Ticker + "-" + hex.EncodeToString(key[:3])

		if usedBy := tx.Get("assets:tickers:used:" + tokenIdentification); usedBy != nil {
			return errors.New("tokenIdentification already! Try again")
		}

		tx.Put("assets:tickers:by:"+string(key), []byte(tokenIdentification))
		tx.Put("assets:tickers:used:"+tokenIdentification, []byte{1})

		return
	}

	assets.HashMap.DeletedEvent = func(key []byte) (err error) {
		if !tx.IsWritable() {
			return
		}

		tokenIdentification := tx.Get("assets:tickers:by:" + string(key))

		tx.Delete("assets:tickers:by:" + string(key))
		tx.Delete("assets:tickers:used:" + string(tokenIdentification))
		return
	}

	return
}
