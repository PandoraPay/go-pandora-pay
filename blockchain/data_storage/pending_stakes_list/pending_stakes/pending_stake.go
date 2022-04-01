package pending_stakes

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type PendingStake struct {
	helpers.SerializableInterface `json:"-"  msgpack:"-"`
	PublicKey                     []byte `json:"publicKey" msgpack:"publicKey"`
	PendingAmount                 uint64 `json:"balance" msgpack:"balance"`
	PendingType                   bool   `json:"pendingType" msgpack:"pendingType"`
}

func (d *PendingStake) Validate() error {
	if len(d.PublicKey) != cryptography.PublicKeySize {
		return errors.New("PendingStake PublicKey size is invalid")
	}
	return nil
}

func (d *PendingStake) Serialize(w *helpers.BufferWriter) {
	w.Write(d.PublicKey)
	w.WriteUvarint(d.PendingAmount)
	w.WriteBool(d.PendingType)
}

func (d *PendingStake) Deserialize(r *helpers.BufferReader) (err error) {
	if d.PublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
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
