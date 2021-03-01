package transaction_simple

import "pandora-pay/helpers"

type TransactionSimpleOutput struct {
	PublicKeyHash [20]byte
	Amount        uint64
	Token         []byte
}

func (vout *TransactionSimpleOutput) Serialize(writer *helpers.BufferWriter) {
	writer.Write(vout.PublicKeyHash[:])
	writer.WriteUint64(vout.Amount)
	writer.WriteToken(vout.Token[:])
}

func (vout *TransactionSimpleOutput) Deserialize(reader *helpers.BufferReader) (err error) {
	if vout.PublicKeyHash, err = reader.Read20(); err != nil {
		return
	}
	if vout.Amount, err = reader.ReadUvarint(); err != nil {
		return
	}
	if vout.Token, err = reader.ReadToken(); err != nil {
		return
	}
	return
}
