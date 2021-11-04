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
	changes                       []*AssetFeeLiquidity
}

type UpdateLiquidityStatus byte

const (
	UPDATE_LIQUIDITY_NOTHING = iota
	UPDATE_LIQUIDITY_OVERWRITTEN
	UPDATE_LIQUIDITY_INSERTED
	UPDATE_LIQUIDITY_DELETED
)

func (self *AssetFeeLiquidities) UpdateLiquidity(updateLiquidity *AssetFeeLiquidity) (UpdateLiquidityStatus, error) {

	if updateLiquidity.ConversionRate == 0 {
		for i, it := range self.List {
			if bytes.Equal(it.AssetId, updateLiquidity.AssetId) {
				self.List = append(self.List[:i], self.List[i+1:]...)
				return UPDATE_LIQUIDITY_DELETED, nil
			}
		}
		return UPDATE_LIQUIDITY_NOTHING, nil
	} else {
		for _, it := range self.List {
			if bytes.Equal(it.AssetId, updateLiquidity.AssetId) {
				it.ConversionRate = updateLiquidity.ConversionRate
				return UPDATE_LIQUIDITY_OVERWRITTEN, nil
			}
		}
		if len(self.List) > 255 {
			return 0, errors.New("AssetFeeLiquidityList will exceed the max")
		}
		self.List = append(self.List, &AssetFeeLiquidity{
			AssetId:        updateLiquidity.AssetId,
			ConversionRate: updateLiquidity.ConversionRate,
		})
		return UPDATE_LIQUIDITY_INSERTED, nil
	}

}

func (self *AssetFeeLiquidities) Serialize(w *helpers.BufferWriter) {
	w.WriteByte(byte(len(self.List)))
	if len(self.List) > 0 {
		w.WriteUvarint(uint64(self.Version))
		for _, liquidity := range self.List {
			liquidity.Serialize(w)
		}
	}
}

func (self *AssetFeeLiquidities) Deserialize(r *helpers.BufferReader) (err error) {

	var count byte
	if count, err = r.ReadByte(); err != nil {
		return
	}

	if count > 0 {
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

		self.List = make([]*AssetFeeLiquidity, count)
		for _, item := range self.List {
			if err = item.Deserialize(r); err != nil {
				return
			}
		}
	}

	return
}
