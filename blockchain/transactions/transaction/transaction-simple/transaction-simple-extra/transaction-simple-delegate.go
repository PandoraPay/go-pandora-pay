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
	Amount       uint64
	HasNewData   bool
	NewPublicKey helpers.HexBytes //20 byte
	NewFee       uint16
}

func (tx *TransactionSimpleDelegate) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) (err error) {
	if err = acc.AddBalance(false, tx.Amount, config.NATIVE_TOKEN); err != nil {
		return
	}
	if !acc.HasDelegatedStake() {
		if err = acc.CreateDelegatedStake(0, tx.NewPublicKey, tx.NewFee); err != nil {
			return
		}
	}
	if err = acc.DelegatedStake.AddStakePendingStake(tx.Amount, blockHeight); err != nil {
		return
	}
	if tx.HasNewData {
		acc.DelegatedStake.DelegatedPublicKey = tx.NewPublicKey
	}
	return
}

func (tx *TransactionSimpleDelegate) Validate() error {

	if tx.HasNewData {

		if len(tx.NewPublicKey) != cryptography.PublicKeySize {
			return errors.New("New Public Key Hash length is invalid")
		}

	} else {
		if len(tx.NewPublicKey) != 0 || tx.NewFee != 0 {
			return errors.New("New Public Key Hash and Fee must be empty")
		}
		if tx.Amount == 0 {
			return errors.New("Transaction Delegate arguments are empty")
		}
	}
	return nil
}

func (tx *TransactionSimpleDelegate) Serialize(writer *helpers.BufferWriter) {
	writer.WriteUvarint(tx.Amount)
	writer.WriteBool(tx.HasNewData)
	if tx.HasNewData {
		writer.Write(tx.NewPublicKey)
		writer.WriteUvarint16(tx.NewFee)
	}
}

func (tx *TransactionSimpleDelegate) Deserialize(reader *helpers.BufferReader) (err error) {
	if tx.Amount, err = reader.ReadUvarint(); err != nil {
		return
	}
	if tx.HasNewData, err = reader.ReadBool(); err != nil {
		return
	}
	if tx.HasNewData {
		if tx.NewPublicKey, err = reader.ReadBytes(cryptography.PublicKeySize); err != nil {
			return
		}
		if tx.NewFee, err = reader.ReadUvarint16(); err != nil {
			return
		}
	}
	return
}
