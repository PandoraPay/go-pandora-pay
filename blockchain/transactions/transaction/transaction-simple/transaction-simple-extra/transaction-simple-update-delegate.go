package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleUpdateDelegate struct {
	TransactionSimpleExtraInterface
	NewPublicKey helpers.HexBytes //20 byte
	NewFee       uint16
}

func (tx *TransactionSimpleUpdateDelegate) IncludeTransactionVin0(blockHeight uint64, acc *account.Account) (err error) {
	if !acc.HasDelegatedStake() {
		if err = acc.CreateDelegatedStake(0, tx.NewPublicKey, tx.NewFee); err != nil {
			return
		}
	} else {
		acc.DelegatedStake.DelegatedPublicKey = tx.NewPublicKey
		acc.DelegatedStake.DelegatedStakeFee = tx.NewFee
	}
	return
}

func (tx *TransactionSimpleUpdateDelegate) Validate() error {
	if len(tx.NewPublicKey) != cryptography.PublicKeySize {
		return errors.New("New Public Key Hash length is invalid")
	}
	return nil
}

func (tx *TransactionSimpleUpdateDelegate) Serialize(writer *helpers.BufferWriter) {
	writer.Write(tx.NewPublicKey)
	writer.WriteUvarint16(tx.NewFee)
}

func (tx *TransactionSimpleUpdateDelegate) Deserialize(reader *helpers.BufferReader) (err error) {
	if tx.NewPublicKey, err = reader.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if tx.NewFee, err = reader.ReadUvarint16(); err != nil {
		return
	}
	return
}
