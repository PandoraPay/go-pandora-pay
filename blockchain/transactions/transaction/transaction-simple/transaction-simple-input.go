package transaction_simple

import (
	"pandora-pay/helpers"
)

type TransactionSimpleInput struct {
	Amount    uint64
	Token     helpers.HexBytes             //20
	Signature helpers.HexBytes             //65
	Bloom     *TransactionSimpleInputBloom `json:"-"`
}

func (vin *TransactionSimpleInput) Serialize(writer *helpers.BufferWriter, inclSignature bool) {
	writer.WriteUvarint(vin.Amount)
	writer.WriteToken(vin.Token)
	if inclSignature {
		writer.Write(vin.Signature)
	}
}

func (vin *TransactionSimpleInput) Deserialize(reader *helpers.BufferReader) (err error) {

	if vin.Amount, err = reader.ReadUvarint(); err != nil {
		return err
	}
	if vin.Token, err = reader.ReadToken(); err != nil {
		return err
	}
	if vin.Signature, err = reader.ReadBytes(65); err != nil {
		return err
	}
	return
}
