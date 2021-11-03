package asset_fee_liquidity

import (
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
)

type AssetFeeLiquidity struct {
	helpers.SerializableInterface `json:"-"`
	AssetId                       helpers.HexBytes `json:"assetId"`
	ConversionRate                uint64           `json:"conversionRate"`
}

func (plainAccount *AssetFeeLiquidity) Serialize(w *helpers.BufferWriter) {
	w.Write(plainAccount.AssetId)
	w.WriteUvarint(plainAccount.ConversionRate)
}

func (plainAccount *AssetFeeLiquidity) Deserialize(r *helpers.BufferReader) (err error) {
	if plainAccount.AssetId, err = r.ReadBytes(config_coins.ASSET_LENGTH); err != nil {
		return
	}
	if plainAccount.ConversionRate, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}
