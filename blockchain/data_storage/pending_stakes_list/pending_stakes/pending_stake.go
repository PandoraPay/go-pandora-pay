package pending_stakes

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type PendingStake struct {
	PublicKey     []byte `json:"publicKey" msgpack:"publicKey"`
	PendingAmount []byte `json:"pendingAmount" msgpack:"pendingAmount"`
}

func (d *PendingStake) Validate() error {
	if len(d.PublicKey) != cryptography.PublicKeySize {
		return errors.New("PendingStake PublicKey size is invalid")
	}
	return nil
}

func (d *PendingStake) Serialize(w *helpers.BufferWriter) {
	w.Write(d.PublicKey)
	w.Write(d.PendingAmount)
}

func (d *PendingStake) Deserialize(r *helpers.BufferReader) (err error) {
	if d.PublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if d.PendingAmount, err = r.ReadBytes(66); err != nil {
		return
	}
	return
}
