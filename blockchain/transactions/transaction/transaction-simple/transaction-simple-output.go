package transaction_simple

import "pandora-pay/helpers"

type TransactionSimpleOutput struct {
	PublicKeyHash [20]byte
	Amount        uint64
	Token         [20]byte
}

func (vout *TransactionSimpleOutput) Serialize(writer *helpers.BufferWriter) {
	writer.Write(vout.PublicKeyHash[:])
	writer.WriteUvarint(vout.Amount)
	writer.WriteToken(&vout.Token)
}

func (vout *TransactionSimpleOutput) Deserialize(reader *helpers.BufferReader) {
	vout.PublicKeyHash = reader.Read20()
	vout.Amount = reader.ReadUvarint()
	vout.Token = reader.ReadToken()
}
