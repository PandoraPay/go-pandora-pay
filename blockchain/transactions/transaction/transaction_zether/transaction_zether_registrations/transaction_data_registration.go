package transaction_zether_registrations

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionZetherDataRegistration struct {
	PublicKeyIndex        uint64
	RegistrationSignature []byte
}

func (registration *TransactionZetherDataRegistration) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(registration.PublicKeyIndex)
	w.Write(registration.RegistrationSignature)
}

func (registration *TransactionZetherDataRegistration) Deserialize(r *helpers.BufferReader) (err error) {
	if registration.PublicKeyIndex, err = r.ReadUvarint(); err != nil {
		return
	}
	if registration.RegistrationSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
