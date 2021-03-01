package transaction_simple

import "pandora-pay/helpers"

type TransactionSimple struct {
	Nonce uint64
	Vin   []*TransactionSimpleInput
	Vout  []*TransactionSimpleOutput
}

func (tx *TransactionSimple) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUint64(tx.Nonce)
	writer.WriteUint64(uint64(len(tx.Vin)))
	for _, vin := range tx.Vin {
		vin.Serialize(writer)
	}
	writer.WriteUint64(uint64(len(tx.Vout)))
	for _, vout := range tx.Vout {
		vout.Serialize(writer)
	}
}

func (tx *TransactionSimple) Deserialize(reader *helpers.BufferReader) (err error) {
	if tx.Nonce, err = reader.ReadUvarint(); err != nil {
		return
	}
	var n uint64
	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	for i := 0; i < n; i++ {
		vin := new(TransactionSimpleInput)
		if err = vin.Deserialize(reader); err != nil {
			return
		}
		tx.Vin = append(tx.Vin, vin)
	}

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	for i := 0; i < n; i++ {
		vout := new(TransactionSimpleOutput)
		if err = vout.Deserialize(reader); vout != nil {
			return
		}
		tx.Vout = append(tx.Vout, vout)
	}

	return
}
