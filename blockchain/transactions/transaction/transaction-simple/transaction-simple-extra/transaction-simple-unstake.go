package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
)

/**
Substracting Amount and FeeExtra from the StakeAvailable
Creating a Unstake Pending
*/
type TransactionSimpleUnstake struct {
	TransactionSimpleExtraInterface
	Amount uint64
}

func (tx *TransactionSimpleUnstake) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) (err error) {
	if !acc.HasDelegatedStake() {
		return errors.New("acc.HasDelegatedStake is null")
	}
	if err = acc.DelegatedStake.AddStakeAvailable(false, tx.Amount); err != nil {
		return
	}
	if err = acc.DelegatedStake.AddStakePendingUnstake(tx.Amount, blockHeight); err != nil {
		return
	}
	return
}

func (tx *TransactionSimpleUnstake) Validate() error {
	if tx.Amount == 0 {
		return errors.New("Unstake must be greater than zero")
	}
	return nil
}

func (tx *TransactionSimpleUnstake) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.Amount)
}

func (tx *TransactionSimpleUnstake) Deserialize(reader *helpers.BufferReader) (err error) {
	if tx.Amount, err = reader.ReadUvarint(); err != nil {
		return
	}
	return
}
