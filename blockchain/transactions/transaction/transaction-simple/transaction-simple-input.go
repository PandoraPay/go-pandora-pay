package transaction_simple

import (
	"pandora-pay/helpers"
)

type TransactionSimpleInput struct {
	PublicKey [33]byte
	Amount    uint64
	Signature [65]byte
	Token     []byte
}

func (vin *TransactionSimpleInput) Serialize(writer *helpers.BufferWriter, inclSignature bool) {
	writer.Write(vin.PublicKey[:])
	writer.WriteUint64(vin.Amount)
	if inclSignature {
		writer.Write(vin.Signature[:])
	}
	writer.WriteToken(vin.Token)
}

func (vin *TransactionSimpleInput) Deserialize(reader *helpers.BufferReader) (err error) {

	if vin.PublicKey, err = reader.Read33(); err != nil {
		return
	}
	if vin.Amount, err = reader.ReadUvarint(); err != nil {
		return
	}
	if vin.Signature, err = reader.Read65(); err != nil {
		return
	}
	if vin.Token, err = reader.ReadToken(); err != nil {
		return
	}

	return
}
