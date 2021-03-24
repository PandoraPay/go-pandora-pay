package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
)

type TransactionSimpleDelegate struct {
	Amount              uint64
	HasNewPublicKeyHash bool
	NewPublicKeyHash    helpers.ByteString //20 byte
}

func (tx *TransactionSimpleDelegate) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) (err error) {
	if err = acc.DelegatedStake.AddStakePendingStake(tx.Amount, blockHeight); err != nil {
		return
	}
	if tx.HasNewPublicKeyHash {
		acc.DelegatedStake.DelegatedPublicKeyHash = tx.NewPublicKeyHash
	}
	return
}

func (tx *TransactionSimpleDelegate) Validate() error {
	if tx.Amount == 0 {
		return errors.New("Amount must be greather than zero")
	}
	return nil
}

func (tx *TransactionSimpleDelegate) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.Amount)
	writer.WriteBool(tx.HasNewPublicKeyHash)
	if tx.HasNewPublicKeyHash {
		writer.Write(tx.NewPublicKeyHash)
	}
}

func (tx *TransactionSimpleDelegate) Deserialize(reader *helpers.BufferReader) (err error) {
	if tx.Amount, err = reader.ReadUvarint(); err != nil {
		return
	}
	if tx.HasNewPublicKeyHash, err = reader.ReadBool(); err != nil {
		return
	}
	if tx.HasNewPublicKeyHash {
		if tx.NewPublicKeyHash, err = reader.ReadBytes(20); err != nil {
			return
		}
	}
	return
}
