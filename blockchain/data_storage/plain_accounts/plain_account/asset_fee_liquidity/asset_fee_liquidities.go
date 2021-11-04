package asset_fee_liquidity

import (
	"bytes"
	"errors"
	"pandora-pay/helpers"
)

type AssetFeeLiquidities struct {
	helpers.SerializableInterface `json:"-"`
	Version                       AssetFeeLiquiditiesVersion `json:"version"`
	List                          []*AssetFeeLiquidity       `json:"list"`
}

func (self *AssetFeeLiquidities) UpdateLiquidity(updateLiquidity *AssetFeeLiquidity) {
	if updateLiquidity.ConversionRate == 0 {
		for i, it := range self.List {
			if bytes.Equal(it.AssetId, updateLiquidity.AssetId) {
				self.List = append(self.List[:i], self.List[i+1:]...)
			}
		}
	} else {
		found := false
		for _, it := range self.List {
			if bytes.Equal(it.AssetId, updateLiquidity.AssetId) {
				it.ConversionRate = updateLiquidity.ConversionRate
				found = true
				break
			}
		}
		if !found {
			self.List = append(self.List, &AssetFeeLiquidity{
				AssetId:        updateLiquidity.AssetId,
				ConversionRate: updateLiquidity.ConversionRate,
			})
		}
	}
}

func (self *AssetFeeLiquidities) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(uint64(self.Version))
	w.WriteByte(byte(len(self.List)))
	for _, liquidity := range self.List {
		liquidity.Serialize(w)
	}
}

func (self *AssetFeeLiquidities) Deserialize(r *helpers.BufferReader) (err error) {
	var n uint64

	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	self.Version = AssetFeeLiquiditiesVersion(n)

	switch self.Version {
	case SIMPLE:
	default:
		return errors.New("Invalid Version")
	}

	var count byte
	if count, err = r.ReadByte(); err != nil {
		return
	}

	self.List = make([]*AssetFeeLiquidity, count)
	for _, item := range self.List {
		if err = item.Deserialize(r); err != nil {
			return
		}
	}
	return
}
