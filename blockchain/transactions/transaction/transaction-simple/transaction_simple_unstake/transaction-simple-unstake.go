package transaction_simple_unstake

import "pandora-pay/helpers"

type TransactionSimpleUnstake struct {
	UnstakeAmount uint64
}

func (tx *TransactionSimpleUnstake) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.UnstakeAmount)
}

func (tx *TransactionSimpleUnstake) Deserialize(reader *helpers.BufferReader) (err error) {

	if tx.UnstakeAmount, err = reader.ReadUvarint(); err != nil {
		return
	}

	return
}
