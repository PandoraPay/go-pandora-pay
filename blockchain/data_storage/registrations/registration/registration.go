package registration

import (
	"errors"
	"pandora-pay/helpers"
	"pandora-pay/store/hash_map"
)

type Registration struct {
	hash_map.HashMapElementSerializableInterface `json:"-"`
	PublicKey                                    []byte `json:"-"` //hashMap key
	Index                                        uint64 `json:"-"` //hashMap index
	Version                                      uint64 `json:"version"`
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
	return nil
}

func (registration *Registration) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(registration.Version)
}

func (registration *Registration) Deserialize(r *helpers.BufferReader) (err error) {
	if registration.Version, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}

func NewRegistration(publicKey []byte, index uint64) *Registration {
	return &Registration{
		PublicKey: publicKey,
		Index:     index,
		Version:   0,
	}
}
