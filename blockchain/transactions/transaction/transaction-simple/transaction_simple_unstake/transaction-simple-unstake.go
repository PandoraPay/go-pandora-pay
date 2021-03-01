package transaction_simple_unstake

import "pandora-pay/helpers"

type TransactionSimpleUnstake struct {
	Fee uint64
}

func (tx *TransactionSimpleUnstake) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUint64(tx.Fee)
}

func (tx *TransactionSimpleUnstake) Deserialize(reader *helpers.BufferReader) (err error) {

	if tx.Fee, err = reader.ReadUvarint(); err != nil {
		return
	}

	return
}
