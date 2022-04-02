package account

import (
	"errors"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type Account struct {
	hash_map.HashMapElementSerializableInterface `json:"-" msgpack:"-"`
	Key                                          []byte `json:"-" msgpack:"-"` //hashmap key
	Asset                                        []byte `json:"-" msgpack:"-"` //collection asset
	Index                                        uint64 `json:"-" msgpack:"-"` //hashmap Index
	Version                                      uint64 `json:"version" msgpack:"version"`
	Balance                                      uint64 `json:"balance" msgpack:"balance"`
}

func (account *Account) IsDeletable() bool {
	return false
}

func (account *Account) SetKey(key []byte) {
	account.Key = key
}

func (account *Account) SetIndex(value uint64) {
	account.Index = value
}

func (account *Account) GetIndex() uint64 {
	return account.Index
}

func (account *Account) Validate() error {
	if account.Version != 0 {
		return errors.New("Version is invalid")
	}
	return nil
}

func (account *Account) AddBalance(sign bool, value uint64) error {
	return helpers.SafeUint64Update(sign, &account.Balance, value)
}

func (account *Account) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(account.Version)
	w.WriteUvarint(account.Balance)
}

func (account *Account) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}
	if n != 0 {
		return errors.New("Invalid Account Version")
	}

	account.Version = n

	if account.Balance, err = r.ReadUvarint(); err != nil {
		return
	}

	return
}

func NewAccount(key []byte, index uint64, asset []byte) (*Account, error) {

	acc := &Account{
		Key:     key,
		Version: 0,
		Asset:   asset,
		Index:   index,
		Balance: 0,
	}

	return acc, nil
}

func NewAccountClear(key []byte, index uint64, asset []byte) (*Account, error) {
	acc := &Account{
		Key:     key,
		Version: 0,
		Asset:   asset,
		Index:   index,
		Balance: 0,
	}

	return acc, nil
}
