package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
)

type TransactionSimpleUnstake struct {
	Amount   uint64
	FeeExtra uint64 //this will be subtracted StakeAvailable
}

func (tx *TransactionSimpleUnstake) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) {
	acc.DelegatedStake.AddStakeAvailable(false, tx.Amount)
	acc.DelegatedStake.AddStakeAvailable(false, tx.FeeExtra)
	acc.DelegatedStake.AddStakePendingUnstake(tx.Amount, blockHeight)
}

func (tx *TransactionSimpleUnstake) Validate() error {
	if tx.Amount == 0 {
		return errors.New("Unstake must be greather than zero")
	}
	return nil
}

func (tx *TransactionSimpleUnstake) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.Amount)
	writer.WriteUvarint(tx.FeeExtra)
}

func (tx *TransactionSimpleUnstake) Deserialize(reader *helpers.BufferReader) (err error) {
	if tx.Amount, err = reader.ReadUvarint(); err != nil {
		return
	}
	if tx.FeeExtra, err = reader.ReadUvarint(); err != nil {
		return
	}
	return
}
