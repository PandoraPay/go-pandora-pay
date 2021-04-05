package transaction_simple

import "pandora-pay/helpers"

type TransactionSimpleOutput struct {
	PublicKeyHash helpers.HexBytes //20
	Amount        uint64
	Token         helpers.HexBytes //20
}

func (vout *TransactionSimpleOutput) Serialize(writer *helpers.BufferWriter) {
	writer.Write(vout.PublicKeyHash)
	writer.WriteUvarint(vout.Amount)
	writer.WriteToken(vout.Token)
}

func (vout *TransactionSimpleOutput) Deserialize(reader *helpers.BufferReader) (err error) {
	if vout.PublicKeyHash, err = reader.ReadBytes(20); err != nil {
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
