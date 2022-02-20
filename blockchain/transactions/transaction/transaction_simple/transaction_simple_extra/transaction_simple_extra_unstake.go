package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/helpers"
)

/**
Substracting Amount from the StakeAvailable
Creating a Unstake Pending
*/
type TransactionSimpleExtraUnstake struct {
	TransactionSimpleExtraInterface
	Amount uint64
}

func (txExtra *TransactionSimpleExtraUnstake) IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {
	//if plainAcc == nil {
	//	return errors.New("acc is null")
	//}
	//if err = plainAcc.DelegatedStake.AddStakeAvailable(false, txExtra.Amount); err != nil {
	//	return
	//}
	//if err = plainAcc.DelegatedStake.AddStakePendingUnstake(txExtra.Amount, blockHeight); err != nil {
	//	return
	//}
	return
}

func (txExtra *TransactionSimpleExtraUnstake) Validate() error {
	if txExtra.Amount == 0 {
		return errors.New("Unstake must be greater than zero")
	}
	return nil
}

func (txExtra *TransactionSimpleExtraUnstake) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(txExtra.Amount)
}

func (txExtra *TransactionSimpleExtraUnstake) Deserialize(r *helpers.BufferReader) (err error) {
	if txExtra.Amount, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}
