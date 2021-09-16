package transaction_zether

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionZetherRegistration struct {
	PublicKeyIndex        uint64
	RegistrationSignature []byte
}

func (registration *TransactionZetherRegistration) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(registration.PublicKeyIndex)
	w.Write(registration.RegistrationSignature)
}

func (registration *TransactionZetherRegistration) Deserialize(r *helpers.BufferReader) (err error) {
	if registration.PublicKeyIndex, err = r.ReadUvarint(); err != nil {
		return
	}
	if registration.RegistrationSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
