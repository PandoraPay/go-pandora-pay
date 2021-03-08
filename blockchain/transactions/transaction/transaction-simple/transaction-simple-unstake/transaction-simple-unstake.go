package transaction_simple_unstake

import (
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
)

type TransactionSimpleUnstake struct {
	UnstakeAmount   uint64
	UnstakeFeeExtra uint64
}

func (tx *TransactionSimpleUnstake) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) {
	acc.DelegatedStake.AddStakeAvailable(false, tx.UnstakeAmount)
	acc.DelegatedStake.AddStakeAvailable(false, tx.UnstakeFeeExtra)
	acc.DelegatedStake.AddUnstakeAmount(true, tx.UnstakeAmount, blockHeight)
}

func (tx *TransactionSimpleUnstake) RemoveTransactionVin0(blockHeight uint64, acc *account.Account) {
	acc.DelegatedStake.AddUnstakeAmount(false, tx.UnstakeAmount, blockHeight)
	acc.DelegatedStake.AddStakeAvailable(true, tx.UnstakeFeeExtra)
	acc.DelegatedStake.AddStakeAvailable(true, tx.UnstakeAmount)
}

func (tx *TransactionSimpleUnstake) Validate() {
	if tx.UnstakeAmount == 0 {
		panic("Unstake must be greather than zero")
	}
}

func (tx *TransactionSimpleUnstake) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.UnstakeAmount)
	writer.WriteUvarint(tx.UnstakeFeeExtra)
}

func (tx *TransactionSimpleUnstake) Deserialize(reader *helpers.BufferReader) {
	tx.UnstakeAmount = reader.ReadUvarint()
	tx.UnstakeFeeExtra = reader.ReadUvarint()
}
