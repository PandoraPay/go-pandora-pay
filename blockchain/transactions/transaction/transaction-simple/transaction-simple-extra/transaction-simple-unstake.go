package transaction_simple_extra

import (
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

func (tx *TransactionSimpleUnstake) Validate() {
	if tx.Amount == 0 {
		panic("Unstake must be greather than zero")
	}
}

func (tx *TransactionSimpleUnstake) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.Amount)
	writer.WriteUvarint(tx.FeeExtra)
}

func (tx *TransactionSimpleUnstake) Deserialize(reader *helpers.BufferReader) {
	tx.Amount = reader.ReadUvarint()
	tx.FeeExtra = reader.ReadUvarint()
}
