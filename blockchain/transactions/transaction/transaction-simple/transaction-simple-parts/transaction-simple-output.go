package transaction_simple_parts

import (
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleOutput struct {
	Amount                uint64
	PublicKey             []byte
	HasRegistration       bool
	RegistrationSignature []byte
}

func (vout *TransactionSimpleOutput) Validate() error {
	if vout.Amount == 0 {
		return errors.New("Amount must be greater than zero")
	}
	if len(vout.PublicKey) != cryptography.PublicKeySize {
		return errors.New("PublicKey length is invalid")
	}

	if (vout.HasRegistration && len(vout.RegistrationSignature) != cryptography.SignatureSize) ||
		(!vout.HasRegistration && len(vout.RegistrationSignature) != 0) {
		return errors.New("RegistrationSignature length is invalid")
	}
	return nil
}

func (vout *TransactionSimpleOutput) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(vout.Amount)
	w.Write(vout.PublicKey)
	w.WriteBool(vout.HasRegistration)
	if vout.HasRegistration {
		w.Write(vout.RegistrationSignature)
	}
}

func (vout *TransactionSimpleOutput) Deserialize(r *helpers.BufferReader) (err error) {
	if vout.Amount, err = r.ReadUvarint(); err != nil {
		return
	}
	if vout.PublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if vout.HasRegistration, err = r.ReadBool(); err != nil {
		return
	}
	if vout.HasRegistration {
		if vout.RegistrationSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
			return
		}
	}
	return
}
