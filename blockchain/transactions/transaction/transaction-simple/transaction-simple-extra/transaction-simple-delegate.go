package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleDelegate struct {
	TransactionSimpleExtraInterface
	Amount              uint64
	HasNewPublicKeyHash bool
	NewPublicKeyHash    helpers.HexBytes //20 byte
}

func (tx *TransactionSimpleDelegate) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) (err error) {
	if err = acc.AddBalance(false, tx.Amount, config.NATIVE_TOKEN); err != nil {
		return
	}
	if !acc.HasDelegatedStake() {
		if err = acc.CreateDelegatedStake(0, tx.NewPublicKeyHash); err != nil {
			return
		}
	}
	if err = acc.DelegatedStake.AddStakePendingStake(tx.Amount, blockHeight); err != nil {
		return
	}
	if tx.HasNewPublicKeyHash {
		acc.DelegatedStake.DelegatedPublicKeyHash = tx.NewPublicKeyHash
	}
	return
}

func (tx *TransactionSimpleDelegate) Validate() error {
	if tx.HasNewPublicKeyHash && len(tx.NewPublicKeyHash) != cryptography.KeyHashSize {
		return errors.New("New Public Key Hash length is invalid")
	}
	if !tx.HasNewPublicKeyHash && len(tx.NewPublicKeyHash) != 0 {
		return errors.New("New Public Key Hash length is invalid")
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
