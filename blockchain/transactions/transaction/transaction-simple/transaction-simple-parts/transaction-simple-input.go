package transaction_simple_parts

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleInput struct {
	PublicKey helpers.HexBytes //33
	Signature helpers.HexBytes //64
}

func (vin *TransactionSimpleInput) Serialize(writer *helpers.BufferWriter, inclSignature bool) {
	writer.Write(vin.PublicKey)
	if inclSignature {
		writer.Write(vin.Signature)
	}
}

func (vin *TransactionSimpleInput) Deserialize(reader *helpers.BufferReader) (err error) {
	if vin.PublicKey, err = reader.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if vin.Signature, err = reader.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
