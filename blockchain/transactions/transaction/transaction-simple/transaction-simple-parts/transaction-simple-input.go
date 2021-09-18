package transaction_simple_parts

import (
	"bytes"
	"errors"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleInput struct {
	PublicKey helpers.HexBytes //33
	Signature helpers.HexBytes //64
}

func (vin *TransactionSimpleInput) Validate() error {
	if bytes.Equal(vin.PublicKey, config.BURN_PUBLIC_KEY) {
		return errors.New("Input includes BURN ADDR")
	}

	if len(vin.PublicKey) != cryptography.PublicKeySize {
		return errors.New("Vin.PublicKey length is invalid")
	}
	if len(vin.Signature) != cryptography.SignatureSize {
		return errors.New("Vin.Signature length is invalid")
	}
	return nil
}

func (vin *TransactionSimpleInput) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(vin.PublicKey)
	if inclSignature {
		w.Write(vin.Signature)
	}
}

func (vin *TransactionSimpleInput) Deserialize(r *helpers.BufferReader) (err error) {
	if vin.PublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if vin.Signature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
