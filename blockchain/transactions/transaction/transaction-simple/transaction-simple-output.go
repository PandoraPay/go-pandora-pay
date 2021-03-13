package transaction_simple

import "pandora-pay/helpers"

type TransactionSimpleOutput struct {
	PublicKeyHash helpers.ByteString //20
	Amount        uint64
	Token         helpers.ByteString //20
}

func (vout *TransactionSimpleOutput) Serialize(writer *helpers.BufferWriter) {
	writer.Write(vout.PublicKeyHash)
	writer.WriteUvarint(vout.Amount)
	writer.WriteToken(vout.Token)
}

func (vout *TransactionSimpleOutput) Deserialize(reader *helpers.BufferReader) {
	vout.PublicKeyHash = reader.ReadBytes(20)
	vout.Amount = reader.ReadUvarint()
	vout.Token = reader.ReadToken()
}
