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
	NewFee       uint64
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
	if tx.NewFee > 10000 {
		return errors.New("Invalid NewFee")
	}
	return nil
}

func (tx *TransactionSimpleUpdateDelegate) Serialize(w *helpers.BufferWriter) {
	w.Write(tx.NewPublicKey)
	w.WriteUvarint(tx.NewFee)
}

func (tx *TransactionSimpleUpdateDelegate) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.NewPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if tx.NewFee, err = r.ReadUvarint(); err != nil {
		return
	}
	if tx.NewFee > 10000 {
		return errors.New("Invalid NewFee")
	}
	return
}
