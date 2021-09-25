package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data/accounts"
	"pandora-pay/blockchain/data/accounts/account"
	plain_accounts "pandora-pay/blockchain/data/plain-accounts"
	plain_account "pandora-pay/blockchain/data/plain-accounts/plain-account"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/data/tokens"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction-data"
	"pandora-pay/blockchain/transactions/transaction/transaction-simple/transaction-simple-parts"
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

func (tx *TransactionSimpleClaim) IncludeTransactionVin0(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, plainAcc *plain_account.PlainAccount, regs *registrations.Registrations, plainAccs *plain_accounts.PlainAccounts, accsCollection *accounts.AccountsCollection, toks *tokens.Tokens) (err error) {

	if plainAcc == nil {
		return errors.New("acc.HasDelegatedStake is null")
	}

	publicKeyList := make([][]byte, len(tx.Output))
	for i, out := range tx.Output {
		publicKeyList[i] = out.PublicKey
	}
	if err = txRegistrations.RegisterNow(regs, publicKeyList); err != nil {
		return
	}

	var accs *accounts.Accounts
	if accs, err = accsCollection.GetMap(config.NATIVE_TOKEN); err != nil {
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

		if err = acc.AddBalanceUint(out.Amount); err != nil {
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

func (tx *TransactionSimpleClaim) Serialize(w *helpers.BufferWriter) {
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
