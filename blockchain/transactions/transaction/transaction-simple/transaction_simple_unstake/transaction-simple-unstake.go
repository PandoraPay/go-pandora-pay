package transaction_simple_unstake

import (
	"errors"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/helpers"
)

type TransactionSimpleUnstake struct {
	UnstakeAmount uint64
}

func (tx *TransactionSimpleUnstake) Validate(txType transaction_type.TransactionType) error {
	if tx.UnstakeAmount == 0 {
		return errors.New("Unstake must be greather than zero")
	}
	return nil
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
