package asset_fee_liquidity

import (
	"bytes"
	"errors"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
)

type AssetFeeLiquidity struct {
	helpers.SerializableInterface `json:"-"`
	AssetId                       helpers.HexBytes `json:"assetId"`
	ConversionRate                uint64           `json:"conversionRate"`
}

func (self *AssetFeeLiquidity) Validate() error {
	if len(self.AssetId) != config_coins.ASSET_LENGTH {
		return errors.New("AssetId length is invalid")
	}

	if bytes.Equal(self.AssetId, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("AssetId NATIVE_ASSET_FULL is not allowed")
	}

	return nil
}

func (self *AssetFeeLiquidity) Serialize(w *helpers.BufferWriter) {
	w.Write(self.AssetId)
	w.WriteUvarint(self.ConversionRate)
}

func (self *AssetFeeLiquidity) Deserialize(r *helpers.BufferReader) (err error) {
	if self.AssetId, err = r.ReadBytes(config_coins.ASSET_LENGTH); err != nil {
		return
	}
	if self.ConversionRate, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}
