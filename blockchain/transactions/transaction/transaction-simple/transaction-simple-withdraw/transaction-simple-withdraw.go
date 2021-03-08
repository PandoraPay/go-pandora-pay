package transaction_simple_withdraw

import (
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config"
	"pandora-pay/helpers"
)

type TransactionSimpleWithdraw struct {
	WithdrawAmount   uint64
	WithdrawFeeExtra uint64
}

func (tx *TransactionSimpleWithdraw) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) {
	acc.DelegatedStake.AddStakeAvailable(false, tx.WithdrawAmount)
	acc.AddBalance(true, tx.WithdrawAmount, config.NATIVE_TOKEN)
	acc.DelegatedStake.AddUnstakeAmount(false, tx.WithdrawFeeExtra, blockHeight)
}

func (tx *TransactionSimpleWithdraw) RemoveTransactionVin0(blockHeight uint64, acc *account.Account) {
	acc.DelegatedStake.AddUnstakeAmount(true, tx.WithdrawFeeExtra, blockHeight)
	acc.AddBalance(false, tx.WithdrawAmount, config.NATIVE_TOKEN)
	acc.DelegatedStake.AddStakeAvailable(true, tx.WithdrawAmount)
}

func (tx *TransactionSimpleWithdraw) Validate() {
	if tx.WithdrawAmount == 0 {
		panic("WithdrawAmount must be greather than zero")
	}
}

func (tx *TransactionSimpleWithdraw) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.WithdrawAmount)
	writer.WriteUvarint(tx.WithdrawFeeExtra)
}

func (tx *TransactionSimpleWithdraw) Deserialize(reader *helpers.BufferReader) {
	tx.WithdrawAmount = reader.ReadUvarint()
	tx.WithdrawFeeExtra = reader.ReadUvarint()
}
