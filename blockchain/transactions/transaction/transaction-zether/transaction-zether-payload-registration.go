package transaction_zether

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionZetherRegistration struct {
	Index     uint64
	Signature []byte
}

func (payloadRegistration *TransactionZetherRegistration) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(payloadRegistration.Index)
	w.Write(payloadRegistration.Signature)
}

func (payloadRegistration *TransactionZetherRegistration) Deserialize(r *helpers.BufferReader) (err error) {
	if payloadRegistration.Index, err = r.ReadUvarint(); err != nil {
		return
	}
	if payloadRegistration.Signature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
