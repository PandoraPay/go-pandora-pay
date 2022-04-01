package plain_account

import (
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type PlainAccount struct {
	hash_map.HashMapElementSerializableInterface `json:"-" msgpack:"-"`
	Key                                          []byte `json:"-" msgpack:"-"` //hashMap key
	Index                                        uint64 `json:"-" msgpack:"-"` //hashMap index
	Nonce                                        uint64 `json:"nonce" msgpack:"nonce"`
	Unclaimed                                    uint64 `json:"unclaimed" msgpack:"unclaimed"`
}

func (plainAccount *PlainAccount) IsDeletable() bool {
	if plainAccount.Unclaimed == 0 && plainAccount.Nonce == 0 {
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
	return nil
}

func (plainAccount *PlainAccount) IncrementNonce(sign bool) error {
	return helpers.SafeUint64Update(sign, &plainAccount.Nonce, 1)
}

func (plainAccount *PlainAccount) AddUnclaimed(sign bool, amount uint64) error {
	return helpers.SafeUint64Update(sign, &plainAccount.Unclaimed, amount)
}

func (plainAccount *PlainAccount) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(plainAccount.Nonce)
	w.WriteUvarint(plainAccount.Unclaimed)
}

func (plainAccount *PlainAccount) Deserialize(r *helpers.BufferReader) (err error) {

	if plainAccount.Nonce, err = r.ReadUvarint(); err != nil {
		return
	}
	if plainAccount.Unclaimed, err = r.ReadUvarint(); err != nil {
		return
	}

	return
}

func NewPlainAccount(key []byte, index uint64) *PlainAccount {
	return &PlainAccount{
		Key:   key,
		Index: index,
	}
}
