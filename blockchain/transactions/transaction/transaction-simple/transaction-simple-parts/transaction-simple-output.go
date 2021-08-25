package transaction_simple_parts

import (
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleOutput struct {
	PublicKey helpers.HexBytes //33
	Amount    uint64
}

func (vout *TransactionSimpleOutput) Serialize(writer *helpers.BufferWriter) {
	writer.Write(vout.PublicKey)
	writer.WriteUvarint(vout.Amount)
}

func (vout *TransactionSimpleOutput) Deserialize(reader *helpers.BufferReader) (err error) {
	if vout.PublicKey, err = reader.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if vout.Amount, err = reader.ReadUvarint(); err != nil {
		return
	}
	return
}
