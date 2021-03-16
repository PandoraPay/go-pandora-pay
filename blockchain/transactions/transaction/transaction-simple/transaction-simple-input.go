package transaction_simple

import (
	"pandora-pay/helpers"
)

type TransactionSimpleInput struct {
	Amount    uint64
	Token     helpers.ByteString //20
	Signature helpers.ByteString //65
	Bloom     *TransactionSimpleInputBloom
}

func (vin *TransactionSimpleInput) Serialize(writer *helpers.BufferWriter, inclSignature bool) {
	writer.WriteUvarint(vin.Amount)
	writer.WriteToken(vin.Token)
	if inclSignature {
		writer.Write(vin.Signature)
	}
}

func (vin *TransactionSimpleInput) Deserialize(reader *helpers.BufferReader) {

	vin.Amount = reader.ReadUvarint()
	vin.Signature = reader.ReadBytes(65)
	vin.Token = reader.ReadToken()

}
