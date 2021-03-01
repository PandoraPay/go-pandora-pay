package transaction_simple

import (
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/helpers"
)

type TransactionSimple struct {
	Nonce uint64
	Vin   []*TransactionSimpleInput
	Vout  []*TransactionSimpleOutput
	Extra interface{}
}

func (tx *TransactionSimple) VerifySignature(hash helpers.Hash) bool {
	return false
}

func (tx *TransactionSimple) Serialize(writer *helpers.BufferWriter, inclSignature bool, txType transaction_type.TransactionType) {
	writer.WriteUint64(tx.Nonce)
	writer.WriteUint64(uint64(len(tx.Vin)))
	for _, vin := range tx.Vin {
		vin.Serialize(writer, inclSignature)
	}

	//vout only TransactionTypeSimple
	if txType == transaction_type.TransactionTypeSimple {
		writer.WriteUint64(uint64(len(tx.Vout)))
		for _, vout := range tx.Vout {
			vout.Serialize(writer)
		}
	}
}

func (tx *TransactionSimple) Deserialize(reader *helpers.BufferReader, txType transaction_type.TransactionType) (err error) {

	if tx.Nonce, err = reader.ReadUvarint(); err != nil {
		return
	}
	var n uint64

	if n, err = reader.ReadUvarint(); err != nil {
		return
	}
	for i := 0; i < int(n); i++ {
		vin := new(TransactionSimpleInput)
		if err = vin.Deserialize(reader); err != nil {
			return
		}
		tx.Vin = append(tx.Vin, vin)
	}

	//vout only TransactionTypeSimple
	if txType == transaction_type.TransactionTypeSimple {
		if n, err = reader.ReadUvarint(); err != nil {
			return
		}
		for i := 0; i < int(n); i++ {
			vout := new(TransactionSimpleOutput)
			if err = vout.Deserialize(reader); err != nil {
				return
			}
			tx.Vout = append(tx.Vout, vout)
		}
	}

	return
}
