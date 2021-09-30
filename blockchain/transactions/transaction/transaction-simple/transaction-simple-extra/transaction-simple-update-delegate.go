package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	plain_account "pandora-pay/blockchain/data_storage/plain-accounts/plain-account"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction-data"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type TransactionSimpleUpdateDelegate struct {
	TransactionSimpleExtraInterface
	NewPublicKey helpers.HexBytes //20 byte
	NewFee       uint64
}

func (tx *TransactionSimpleUpdateDelegate) IncludeTransactionVin0(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {

	if len(txRegistrations.Registrations) > 0 {
		return errors.New("txRegistrations.Registrations length should be zero")
	}

	if plainAcc == nil || !plainAcc.HasDelegatedStake() {
		if err = plainAcc.CreateDelegatedStake(0, tx.NewPublicKey, tx.NewFee); err != nil {
			return
		}
	} else {
		plainAcc.DelegatedStake.DelegatedPublicKey = tx.NewPublicKey
		plainAcc.DelegatedStake.DelegatedStakeFee = tx.NewFee
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
