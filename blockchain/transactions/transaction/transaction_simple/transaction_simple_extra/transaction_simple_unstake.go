package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

/**
Substracting Amount from the StakeAvailable
Creating a Unstake Pending
*/
type TransactionSimpleUnstake struct {
	TransactionSimpleExtraInterface
	Amount uint64
}

func (tx *TransactionSimpleUnstake) IncludeTransactionVin0(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {

	if len(txRegistrations.Registrations) > 0 {
		return errors.New("txRegistrations.Registrations length should be zero")
	}

	if plainAcc == nil || !plainAcc.HasDelegatedStake() {
		return errors.New("acc.HasDelegatedStake is null")
	}
	if err = plainAcc.DelegatedStake.AddStakeAvailable(false, tx.Amount); err != nil {
		return
	}
	if err = plainAcc.DelegatedStake.AddStakePendingUnstake(tx.Amount, blockHeight); err != nil {
		return
	}
	return
}

func (tx *TransactionSimpleUnstake) Validate() error {
	if tx.Amount == 0 {
		return errors.New("Unstake must be greater than zero")
	}
	return nil
}

func (tx *TransactionSimpleUnstake) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(tx.Amount)
}

func (tx *TransactionSimpleUnstake) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.Amount, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}
