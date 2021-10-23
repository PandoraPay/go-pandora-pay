package transaction_zether_registrations

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionZetherDataRegistration struct {
	PublicKeyIndex        byte
	RegistrationSignature []byte
}

func (registration *TransactionZetherDataRegistration) Serialize(w *helpers.BufferWriter) {
	w.WriteByte(registration.PublicKeyIndex)
	w.Write(registration.RegistrationSignature)
}

func (registration *TransactionZetherDataRegistration) Deserialize(r *helpers.BufferReader) (err error) {
	if registration.PublicKeyIndex, err = r.ReadByte(); err != nil {
		return
	}
	if registration.RegistrationSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
