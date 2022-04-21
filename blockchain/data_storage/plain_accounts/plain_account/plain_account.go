package plain_account

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account/dpos"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type PlainAccount struct {
	hash_map.HashMapElementSerializableInterface `json:"-" msgpack:"-"`
	Key                                          []byte               `json:"-" msgpack:"-"` //hashMap key
	Index                                        uint64               `json:"-" msgpack:"-"` //hashMap index
	Nonce                                        uint64               `json:"nonce" msgpack:"nonce"`
	StakeAvailable                               uint64               `json:"stakeAvailable,omitempty" msgpack:"stakeAvailable,omitempty"` //confirmed stake
	DelegatedStake                               *dpos.DelegatedStake `json:"delegatedStake" msgpack:"delegatedStake"`
}

func (plainAccount *PlainAccount) IsDeletable() bool {
	return plainAccount.Nonce == 0 && plainAccount.DelegatedStake.IsDeletable()
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
	if err := plainAccount.DelegatedStake.Validate(); err != nil {
		return err
	}
	return nil
}

func (plainAccount *PlainAccount) IncrementNonce(sign bool) error {
	return helpers.SafeUint64Update(sign, &plainAccount.Nonce, 1)
}

func (plainAccount *PlainAccount) AddStakeAvailable(sign bool, amount uint64) error {
	return helpers.SafeUint64Update(sign, &plainAccount.StakeAvailable, amount)
}

func (plainAccount *PlainAccount) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(plainAccount.Nonce)
	w.WriteUvarint(plainAccount.StakeAvailable)
	plainAccount.DelegatedStake.Serialize(w)
}

func (plainAccount *PlainAccount) Deserialize(r *helpers.BufferReader) (err error) {
	if plainAccount.Nonce, err = r.ReadUvarint(); err != nil {
		return
	}
	if plainAccount.StakeAvailable, err = r.ReadUvarint(); err != nil {
		return
	}
	return plainAccount.DelegatedStake.Deserialize(r)
}

func NewPlainAccount(key []byte, index uint64) *PlainAccount {
	return &PlainAccount{
		Key:            key,
		Index:          index,
		DelegatedStake: &dpos.DelegatedStake{Version: dpos.NO_STAKING},
	}
}
