package transaction_simple_extra

import (
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
)

type TransactionSimpleDelegate struct {
	Amount              uint64
	HasNewPublicKeyHash bool
	NewPublicKeyHash    helpers.ByteString //20 byte
}

func (tx *TransactionSimpleDelegate) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) {
	acc.DelegatedStake.AddStakePendingStake(tx.Amount, blockHeight)
	if tx.HasNewPublicKeyHash {
		acc.DelegatedStake.DelegatedPublicKeyHash = tx.NewPublicKeyHash
	}
}

func (tx *TransactionSimpleDelegate) Validate() {
	if tx.Amount == 0 {
		panic("Amount must be greather than zero")
	}
}

func (tx *TransactionSimpleDelegate) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.Amount)
	writer.WriteBool(tx.HasNewPublicKeyHash)
	if tx.HasNewPublicKeyHash {
		writer.Write(tx.NewPublicKeyHash)
	}
}

func (tx *TransactionSimpleDelegate) Deserialize(reader *helpers.BufferReader) {
	tx.Amount = reader.ReadUvarint()
	tx.HasNewPublicKeyHash = reader.ReadBool()
	if tx.HasNewPublicKeyHash {
		tx.NewPublicKeyHash = reader.ReadBytes(20)
	}
}
