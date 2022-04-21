package pending_stakes

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type PendingStake struct {
	helpers.SerializableInterface `json:"-"  msgpack:"-"`
	PublicKeyHash                 []byte `json:"publicKeyHash" msgpack:"publicKeyHash"`
	PendingAmount                 uint64 `json:"balance" msgpack:"balance"`
	PendingType                   bool   `json:"pendingType" msgpack:"pendingType"`
}

func (d *PendingStake) Validate() error {
	if len(d.PublicKeyHash) != cryptography.PublicKeyHashSize {
		return errors.New("PendingStake PublicKey size is invalid")
	}
	return nil
}

func (d *PendingStake) Serialize(w *helpers.BufferWriter) {
	w.Write(d.PublicKeyHash)
	w.WriteUvarint(d.PendingAmount)
	w.WriteBool(d.PendingType)
}

func (d *PendingStake) Deserialize(r *helpers.BufferReader) (err error) {
	if d.PublicKeyHash, err = r.ReadBytes(cryptography.PublicKeyHashSize); err != nil {
		return
	}
	if d.PendingAmount, err = r.ReadUvarint(); err != nil {
		return
	}
	if d.PendingType, err = r.ReadBool(); err != nil {
		return
	}
	return
}
