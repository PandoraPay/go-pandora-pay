package transaction_simple_extra

import (
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
)

type TransactionSimpleDelegate struct {
	DelegateAmount          uint64
	DelegateHasNewPublicKey bool
	DelegateNewPublicKey    [33]byte
}

func (tx *TransactionSimpleDelegate) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) {
	acc.DelegatedStake.AddStakePending(true, tx.DelegateAmount, blockHeight)
	if tx.DelegateHasNewPublicKey {
		acc.DelegatedStake.DelegatedPublicKey = tx.DelegateNewPublicKey
	}
}

func (tx *TransactionSimpleDelegate) RemoveTransactionVin0(blockHeight uint64, acc *account.Account) {
	acc.DelegatedStake.AddStakePending(false, tx.DelegateAmount, blockHeight)
}

func (tx *TransactionSimpleDelegate) Validate() {
	if tx.DelegateAmount == 0 {
		panic("DelegateAmount must be greather than zero")
	}
}

func (tx *TransactionSimpleDelegate) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.DelegateAmount)
	writer.WriteBool(tx.DelegateHasNewPublicKey)
	if tx.DelegateHasNewPublicKey {
		writer.Write(tx.DelegateNewPublicKey[:])
	}
}

func (tx *TransactionSimpleDelegate) Deserialize(reader *helpers.BufferReader) {
	tx.DelegateAmount = reader.ReadUvarint()
	tx.DelegateHasNewPublicKey = reader.ReadBool()
	if tx.DelegateHasNewPublicKey {
		tx.DelegateNewPublicKey = reader.Read33()
	}
}
