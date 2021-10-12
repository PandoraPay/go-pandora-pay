package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/config"
	"pandora-pay/helpers"
)

/**
Substracting Amount and FeeExtra from the Claimable
*/
type TransactionSimpleClaim struct {
	TransactionSimpleExtraInterface
	Output []*transaction_simple_parts.TransactionSimpleOutput
}

func (tx *TransactionSimpleClaim) IncludeTransactionVin0(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) (err error) {

	if plainAcc == nil {
		return errors.New("acc.HasDelegatedStake is null")
	}

	publicKeyList := make([][]byte, len(tx.Output))
	for i, out := range tx.Output {
		publicKeyList[i] = out.PublicKey
	}
	if err = txRegistrations.RegisterNow(dataStorage.Regs, publicKeyList); err != nil {
		return
	}

	var accs *accounts.Accounts
	if accs, err = dataStorage.AccsCollection.GetMap(config.NATIVE_ASSET); err != nil {
		return
	}

	for _, out := range tx.Output {

		if err = plainAcc.AddClaimable(false, out.Amount); err != nil {
			return
		}

		var acc *account.Account
		if acc, err = accs.GetAccount(out.PublicKey); err != nil {
			return
		}
		if acc == nil {
			if acc, err = accs.CreateAccount(out.PublicKey); err != nil {
				return
			}
		}

		if err = acc.Balance.AddBalanceUint(out.Amount); err != nil {
			return
		}

		if err = accs.Update(string(out.PublicKey), acc); err != nil {
			return
		}

	}
	return
}

func (tx *TransactionSimpleClaim) Validate() error {

	if len(tx.Output) == 0 || len(tx.Output) > 255 {
		return errors.New("Clain output length is invalid")
	}

	duplicates := make(map[string]bool)
	for _, out := range tx.Output {
		duplicates[string(out.PublicKey)] = true
		if err := out.Validate(); err != nil {
			return err
		}
	}
	if len(duplicates) != len(tx.Output) {
		return errors.New("Output ")
	}

	return nil
}

func (tx *TransactionSimpleClaim) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.WriteUvarint(uint64(len(tx.Output)))
	for _, out := range tx.Output {
		out.Serialize(w)
	}
}

func (tx *TransactionSimpleClaim) Deserialize(r *helpers.BufferReader) (err error) {

	var n uint64
	if n, err = r.ReadUvarint(); err != nil {
		return
	}

	tx.Output = make([]*transaction_simple_parts.TransactionSimpleOutput, n)
	for i := uint64(0); i < n; i++ {
		tx.Output[i] = &transaction_simple_parts.TransactionSimpleOutput{}
		if err = tx.Output[i].Deserialize(r); err != nil {
			return
		}
	}
	return
}
