package asset_fee_liquidity

import (
	"bytes"
	"errors"
	"pandora-pay/config/config_assets"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
)

type AssetFeeLiquidity struct {
	Asset        []byte `json:"assetId" msgpack:"assetId"`
	Rate         uint64 `json:"rate" msgpack:"rate"`
	LeadingZeros byte   `json:"leadingZeros" msgpack:"leadingZeros"`
}

func (self *AssetFeeLiquidity) Validate() error {
	if len(self.Asset) != config_coins.ASSET_LENGTH {
		return errors.New("AssetId length is invalid")
	}

	if bytes.Equal(self.Asset, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("AssetId NATIVE_ASSET_FULL is not allowed")
	}
	if self.LeadingZeros > config_assets.ASSETS_DECIMAL_SEPARATOR_MAX_BYTE {
		return errors.New("Invalid Leading Zeros")
	}

	return nil
}

func (self *AssetFeeLiquidity) Serialize(w *helpers.BufferWriter) {
	w.Write(self.Asset)
	w.WriteUvarint(self.Rate)
	w.WriteByte(self.LeadingZeros)
}

func (self *AssetFeeLiquidity) Deserialize(r *helpers.BufferReader) (err error) {
	if self.Asset, err = r.ReadBytes(config_coins.ASSET_LENGTH); err != nil {
		return
	}
	if self.Rate, err = r.ReadUvarint(); err != nil {
		return
	}
	if self.LeadingZeros, err = r.ReadByte(); err != nil {
		return
	}
	return
}
