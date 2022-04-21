package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/config/config_stake"
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

func (txExtra *TransactionSimpleExtraUnstake) IncludeTransactionExtra(blockHeight uint64, vinPublicKeyHashes [][]byte, vin []*transaction_simple_parts.TransactionSimpleInput, vout []*transaction_simple_parts.TransactionSimpleOutput, dataStorage *data_storage.DataStorage) (err error) {

	var plainAcc *plain_account.PlainAccount

	for i := range vin {

		if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(vinPublicKeyHashes[i]); err != nil {
			return
		}

		if err = plainAcc.AddStakeAvailable(false, txExtra.Amount); err != nil {
			return
		}

		if err = dataStorage.PlainAccs.Update(string(vinPublicKeyHashes[i]), plainAcc); err != nil {
			return
		}

		if err = dataStorage.AddStakePendingStake(vinPublicKeyHashes[i], txExtra.Amount, false, config_stake.GetPendingStakeWindow(blockHeight)); err != nil {
			return
		}

	}

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
