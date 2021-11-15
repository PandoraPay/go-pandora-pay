package plain_account

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/asset_fee_liquidity"
	dpos "pandora-pay/blockchain/data_storage/plain_accounts/plain_account/dpos"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type PlainAccount struct {
	hash_map.HashMapElementSerializableInterface `json:"-"`
	PublicKey                                    []byte                                   `json:"-"` //hashMap key
	Index                                        uint64                                   `json:"-"` //hashMap index
	Nonce                                        uint64                                   `json:"nonce"`
	Unclaimed                                    uint64                                   `json:"unclaimed"`
	DelegatedStake                               *dpos.DelegatedStake                     `json:"delegatedStake"`
	AssetFeeLiquidities                          *asset_fee_liquidity.AssetFeeLiquidities `json:"assetFeeLiquidities"`
}

func (plainAccount *PlainAccount) SetIndex(value uint64) {
	plainAccount.Index = value
}

func (plainAccount *PlainAccount) GetIndex() uint64 {
	return plainAccount.Index
}

func (plainAccount *PlainAccount) Validate() error {
	if err := plainAccount.DelegatedStake.Validate(); err != nil {
		return err
	}
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

func (plainAccount *PlainAccount) RefreshDelegatedStake(blockHeight uint64) error {

	if !plainAccount.DelegatedStake.HasDelegatedStake() {
		return nil
	}

	unclaimed, err := plainAccount.DelegatedStake.RefreshDelegatedStake(blockHeight)
	if err != nil {
		return err
	}

	return plainAccount.AddUnclaimed(true, unclaimed)
}

func (plainAccount *PlainAccount) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(plainAccount.Nonce)
	w.WriteUvarint(plainAccount.Unclaimed)
	plainAccount.DelegatedStake.Serialize(w)
	plainAccount.AssetFeeLiquidities.Serialize(w)
}

func (plainAccount *PlainAccount) Deserialize(r *helpers.BufferReader) (err error) {

	if plainAccount.Nonce, err = r.ReadUvarint(); err != nil {
		return
	}
	if plainAccount.Unclaimed, err = r.ReadUvarint(); err != nil {
		return
	}

	if err = plainAccount.DelegatedStake.Deserialize(r); err != nil {
		return
	}
	if err = plainAccount.AssetFeeLiquidities.Deserialize(r); err != nil {
		return
	}

	return
}

func NewPlainAccount(publicKey []byte, index uint64) *PlainAccount {
	return &PlainAccount{
		PublicKey:           publicKey,
		Index:               index,
		DelegatedStake:      &dpos.DelegatedStake{},
		AssetFeeLiquidities: &asset_fee_liquidity.AssetFeeLiquidities{},
	}
}
