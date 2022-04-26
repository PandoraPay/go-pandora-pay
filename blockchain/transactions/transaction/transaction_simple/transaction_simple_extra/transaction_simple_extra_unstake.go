package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/helpers"
)

/**
Substracting Amount from the StakeAvailable
Creating a Unstake Pending
*/
type TransactionSimpleExtraUnstake struct {
	TransactionSimpleExtraInterface
	Amounts []uint64
}

func (txExtra *TransactionSimpleExtraUnstake) IncludeTransactionExtra(blockHeight uint64, vinPublicKeyHashes [][]byte, vin []*transaction_simple_parts.TransactionSimpleInput, vout []*transaction_simple_parts.TransactionSimpleOutput, dataStorage *data_storage.DataStorage) (err error) {

	var plainAcc *plain_account.PlainAccount

	for i := range vin {

		if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(vinPublicKeyHashes[i]); err != nil {
			return
		}

		if err = plainAcc.AddStakeAvailable(false, txExtra.Amounts[i]); err != nil {
			return
		}

		if err = dataStorage.PlainAccs.Update(string(vinPublicKeyHashes[i]), plainAcc); err != nil {
			return
		}

		if err = dataStorage.AddStakePendingStake(vinPublicKeyHashes[i], txExtra.Amounts[i], false, blockHeight); err != nil {
			return
		}

	}

	return
}

func (txExtra *TransactionSimpleExtraUnstake) Validate(vin []*transaction_simple_parts.TransactionSimpleInput, vout []*transaction_simple_parts.TransactionSimpleOutput) error {
	if len(vin) != len(txExtra.Amounts) {
		return errors.New("Invalid length")
	}
	for _, amount := range txExtra.Amounts {
		if amount == 0 {
			return errors.New("Unstake must be greater than zero")
		}
	}
	return nil
}

func (txExtra *TransactionSimpleExtraUnstake) Serialize(w *helpers.BufferWriter, vin []*transaction_simple_parts.TransactionSimpleInput, vout []*transaction_simple_parts.TransactionSimpleOutput, inclSignature bool) {
	for _, amount := range txExtra.Amounts {
		w.WriteUvarint(amount)
	}
}

func (txExtra *TransactionSimpleExtraUnstake) Deserialize(r *helpers.BufferReader, vin []*transaction_simple_parts.TransactionSimpleInput, vout []*transaction_simple_parts.TransactionSimpleOutput) (err error) {
	txExtra.Amounts = make([]uint64, len(vin))
	for i := range txExtra.Amounts {
		if txExtra.Amounts[i], err = r.ReadUvarint(); err != nil {
			return
		}
	}
	return
}
