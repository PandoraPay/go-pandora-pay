package transaction_simple_parts

import (
	"bytes"
	"errors"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleInput struct {
	PublicKey []byte //33
	Amount    uint64
	Asset     []byte
	Signature []byte //64
}

func (vin *TransactionSimpleInput) Validate() error {

	if bytes.Equal(vin.PublicKey, config_coins.BURN_PUBLIC_KEY) {
		return errors.New("Input includes BURN ADDR")
	}

	if len(vin.PublicKey) != cryptography.PublicKeySize {
		return errors.New("Vin.PublicKey length is invalid")
	}
	if len(vin.Signature) != cryptography.SignatureSize {
		return errors.New("Vin.Signature length is invalid")
	}
	if len(vin.Asset) != config_coins.ASSET_LENGTH {
		return errors.New("Vin.Asset is invalid")
	}
	return nil
}

func (vin *TransactionSimpleInput) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(vin.PublicKey)
	w.WriteUvarint(vin.Amount)
	w.WriteAsset(vin.Asset)
	if inclSignature {
		w.Write(vin.Signature)
	}
}

func (vin *TransactionSimpleInput) Deserialize(r *helpers.BufferReader) (err error) {
	if vin.PublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if vin.Amount, err = r.ReadUvarint(); err != nil {
		return
	}
	if vin.Asset, err = r.ReadAsset(); err != nil {
		return
	}
	if vin.Signature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
