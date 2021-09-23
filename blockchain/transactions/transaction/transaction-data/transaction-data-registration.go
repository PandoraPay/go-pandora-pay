package transaction_data

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionDataRegistration struct {
	PublicKeyIndex        uint64
	RegistrationSignature []byte
}

func (registration *TransactionDataRegistration) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(registration.PublicKeyIndex)
	w.Write(registration.RegistrationSignature)
}

func (registration *TransactionDataRegistration) Deserialize(r *helpers.BufferReader) (err error) {
	if registration.PublicKeyIndex, err = r.ReadUvarint(); err != nil {
		return
	}
	if registration.RegistrationSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
