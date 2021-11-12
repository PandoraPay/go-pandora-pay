package asset_fee_liquidity

import (
	"bytes"
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type AssetFeeLiquidities struct {
	helpers.SerializableInterface `json:"-"`
	Version                       AssetFeeLiquiditiesVersion `json:"version"`
	List                          []*AssetFeeLiquidity       `json:"list"`
	Collector                     []byte                     `json:"collector"`
}

func (self *AssetFeeLiquidities) HasAssetFeeLiquidities() bool {
	return self.Version == SIMPLE
}

func (self *AssetFeeLiquidities) Clear() {
	self.List = make([]*AssetFeeLiquidity, 0)
	self.Collector = nil
	self.Version = NONE
}

func (self *AssetFeeLiquidities) Validate() error {
	switch self.Version {
	case NONE:
		if len(self.List) != 0 || len(self.Collector) != 0 {
			return errors.New("Collector can not be set while there is no liquidity set")
		}
	case SIMPLE:
		if len(self.List) == 0 || len(self.Collector) != cryptography.PublicKeySize {
			return errors.New("Collector need to be set when there is at least one liquidity provided")
		}
	default:
		return errors.New("Invalid Version")
	}

	return nil
}

func (self *AssetFeeLiquidities) GetLiquidity(assetId []byte) *AssetFeeLiquidity {
	for _, it := range self.List {
		if bytes.Equal(it.AssetId, assetId) {
			return it
		}
	}
	return nil
}

func (self *AssetFeeLiquidities) UpdateLiquidity(updateLiquidity *AssetFeeLiquidity) (UpdateLiquidityStatus, error) {

	if updateLiquidity.Rate == 0 {
		for i, it := range self.List {
			if bytes.Equal(it.AssetId, updateLiquidity.AssetId) {
				self.List = append(self.List[:i], self.List[i+1:]...)

				if len(self.List) == 0 {
					self.Clear()
				}

				return UPDATE_LIQUIDITY_DELETED, nil
			}
		}
		return UPDATE_LIQUIDITY_NOTHING, nil
	} else {
		for _, it := range self.List {
			if bytes.Equal(it.AssetId, updateLiquidity.AssetId) {
				it.Rate = updateLiquidity.Rate
				return UPDATE_LIQUIDITY_OVERWRITTEN, nil
			}
		}
		if len(self.List) > 255 {
			return 0, errors.New("AssetFeeLiquidityList will exceed the max")
		}
		self.List = append(self.List, &AssetFeeLiquidity{
			AssetId: updateLiquidity.AssetId,
			Rate:    updateLiquidity.Rate,
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
		w.Write(self.Collector)
	}
}

func (self *AssetFeeLiquidities) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	self.Version = AssetFeeLiquiditiesVersion(n)

	switch self.Version {
	case NONE:
	case SIMPLE:
		var count byte
		if count, err = r.ReadByte(); err != nil {
			return
		}
		self.List = make([]*AssetFeeLiquidity, count)
		for i := range self.List {
			self.List[i] = &AssetFeeLiquidity{}
			if err = self.List[i].Deserialize(r); err != nil {
				return
			}
		}

		if self.Collector, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
	default:
		return errors.New("Invalid Version")
	}

	return
}
