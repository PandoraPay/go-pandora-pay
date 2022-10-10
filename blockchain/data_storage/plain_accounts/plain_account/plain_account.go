package plain_account

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	"pandora-pay/helpers"
	"pandora-pay/helpers/advanced_buffers"
)

type PlainAccount struct {
	Key                 []byte                                   `json:"-" msgpack:"-"` //hashMap key
	Index               uint64                                   `json:"-" msgpack:"-"` //hashMap index
	Nonce               uint64                                   `json:"nonce" msgpack:"nonce"`
	Unclaimed           uint64                                   `json:"unclaimed" msgpack:"unclaimed"`
	AssetFeeLiquidities *asset_fee_liquidity.AssetFeeLiquidities `json:"assetFeeLiquidities" msgpack:"assetFeeLiquidities"`
}

func (plainAccount *PlainAccount) IsDeletable() bool {
	if plainAccount.Unclaimed == 0 && plainAccount.Nonce == 0 && !plainAccount.AssetFeeLiquidities.HasAssetFeeLiquidities() {
		return true
	}
	return false
}

func (plainAccount *PlainAccount) SetKey(key []byte) {
	plainAccount.Key = key
}

func (plainAccount *PlainAccount) SetIndex(value uint64) {
	plainAccount.Index = value
}

func (plainAccount *PlainAccount) GetIndex() uint64 {
	return plainAccount.Index
}

func (plainAccount *PlainAccount) Validate() error {
	if err := plainAccount.AssetFeeLiquidities.Validate(); err != nil {
		return err
	}
	return nil
}

func (plainAccount *PlainAccount) IncrementNonce(sign bool) error {
	return helpers.SafeUint64Update(sign, &plainAccount.Nonce, 1)
}

func (plainAccount *PlainAccount) AddUnclaimed(sign bool, amount uint64) error {
	return helpers.SafeUint64Update(sign, &plainAccount.Unclaimed, amount)
}

func (plainAccount *PlainAccount) Serialize(w *advanced_buffers.BufferWriter) {
	w.WriteUvarint(plainAccount.Nonce)
	w.WriteUvarint(plainAccount.Unclaimed)
	plainAccount.AssetFeeLiquidities.Serialize(w)
}

func (plainAccount *PlainAccount) Deserialize(r *advanced_buffers.BufferReader) (err error) {

	if plainAccount.Nonce, err = r.ReadUvarint(); err != nil {
		return
	}
	if plainAccount.Unclaimed, err = r.ReadUvarint(); err != nil {
		return
	}
	if err = plainAccount.AssetFeeLiquidities.Deserialize(r); err != nil {
		return
	}

	return
}

func NewPlainAccount(key []byte, index uint64) *PlainAccount {
	return &PlainAccount{
		Key:                 key,
		Index:               index,
		AssetFeeLiquidities: &asset_fee_liquidity.AssetFeeLiquidities{},
	}
}
