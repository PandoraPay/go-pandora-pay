package registration

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type Registration struct {
	hash_map.HashMapElementSerializableInterface `json:"-" msgpack:"-"`
	PublicKey                                    []byte `json:"-" msgpack:"-"` //hashMap key
	Index                                        uint64 `json:"-" msgpack:"-"` //hashMap index
	Version                                      uint64 `json:"version" msgpack:"version"`
	Staked                                       bool   `json:"staked" msgpack:"staked"`
	SpendPublicKey                               []byte `json:"spendPublicKey" msgpack:"spendPublicKey"`
}

func (registration *Registration) SetKey(key []byte) {
	registration.PublicKey = key
}

func (registration *Registration) SetIndex(value uint64) {
	registration.Index = value
}

func (registration *Registration) GetIndex() uint64 {
	return registration.Index
}

func (registration *Registration) Validate() error {
	if registration.Version != 0 {
		return errors.New("Registration Version is invalid")
	}
	if len(registration.SpendPublicKey) != cryptography.PublicKeySize && len(registration.SpendPublicKey) != 0 {
		return errors.New("Spend Public Key length must be 33")
	}
	return nil
}

func (registration *Registration) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(registration.Version)
	w.WriteBool(registration.Staked)
	w.WriteBool(len(registration.SpendPublicKey) > 0)
	w.Write(registration.SpendPublicKey)
}

func (registration *Registration) Deserialize(r *helpers.BufferReader) (err error) {
	if registration.Version, err = r.ReadUvarint(); err != nil {
		return
	}
	if registration.Staked, err = r.ReadBool(); err != nil {
		return
	}

	var hasSpendPublicKey bool
	if hasSpendPublicKey, err = r.ReadBool(); err != nil {
		return
	}
	if hasSpendPublicKey {
		if registration.SpendPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
	}
	return
}

func NewRegistration(publicKey []byte, index uint64) *Registration {
	return &Registration{
		PublicKey:      publicKey,
		Index:          index,
		Version:        0,
		Staked:         false,
		SpendPublicKey: nil,
	}
}
