package assets

import (
	"errors"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/config/config_coins"
	"pandora-pay/store/hash_map"
	"pandora-pay/store/store_db/store_db_interface"
)

type Assets struct {
	*hash_map.HashMap[*asset.Asset]
}

func (this *Assets) CreateAsset(key []byte, ast *asset.Asset) (err error) {

	var exists bool
	if exists, err = this.Exists(string(key)); err != nil {
		return
	}
	if exists {
		return errors.New("Asset already exists")
	}

	return this.Update(string(key), ast)
}

func NewAssets(tx store_db_interface.StoreDBTransactionInterface) (this *Assets) {

	this = &Assets{
		hash_map.CreateNewHashMap[*asset.Asset](tx, "assets", config_coins.ASSET_LENGTH, true),
	}

	this.HashMap.CreateObject = func(key []byte, index uint64) (*asset.Asset, error) {
		return asset.NewAsset(key, index), nil
	}

	this.HashMap.StoredEvent = func(key []byte, committed *hash_map.CommittedMapElement[*asset.Asset], index uint64) (err error) {
		if !this.Tx.IsWritable() {
			return
		}

		if usedBy := this.Tx.Get("assets:tickers:used:" + committed.Element.Identification); usedBy != nil {
			return errors.New("tokenIdentification already! Try again")
		}

		this.Tx.Put("assets:tickers:by:"+string(key), []byte(committed.Element.Identification))
		this.Tx.Put("assets:tickers:used:"+committed.Element.Identification, []byte{1})

		return
	}

	this.HashMap.DeletedEvent = func(key []byte) (err error) {
		if !this.Tx.IsWritable() {
			return
		}

		tokenIdentification := this.Tx.Get("assets:tickers:by:" + string(key))

		this.Tx.Delete("assets:tickers:by:" + string(key))
		this.Tx.Delete("assets:tickers:used:" + string(tokenIdentification))
		return
	}

	return
}
