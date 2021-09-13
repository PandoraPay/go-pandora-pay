package transaction_zether

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionZetherRegistration struct {
	PublicKey             []byte
	RegistrationSignature []byte
}

func (registration *TransactionZetherRegistration) Serialize(w *helpers.BufferWriter) {
	w.Write(registration.PublicKey)
	w.Write(registration.RegistrationSignature)
}

func (registration *TransactionZetherRegistration) Deserialize(r *helpers.BufferReader) (err error) {
	if registration.PublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if registration.RegistrationSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
