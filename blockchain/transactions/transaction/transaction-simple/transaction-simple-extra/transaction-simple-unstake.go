package transaction_simple_extra

import (
	"errors"
	"pandora-pay/blockchain/data/accounts"
	plain_accounts "pandora-pay/blockchain/data/plain-accounts"
	plain_account "pandora-pay/blockchain/data/plain-accounts/plain-account"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/data/tokens"
	"pandora-pay/helpers"
)

/**
Substracting Amount and FeeExtra from the StakeAvailable
Creating a Unstake Pending
*/
type TransactionSimpleUnstake struct {
	TransactionSimpleExtraInterface
	Amount uint64
}

func (tx *TransactionSimpleUnstake) IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, regs *registrations.Registrations, plainAccs *plain_accounts.PlainAccounts, accsCollection *accounts.AccountsCollection, toks *tokens.Tokens) (err error) {
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

func (tx *TransactionSimpleUnstake) Serialize(w *helpers.BufferWriter) {
	w.WriteUvarint(tx.Amount)
}

func (tx *TransactionSimpleUnstake) Deserialize(r *helpers.BufferReader) (err error) {
	if tx.Amount, err = r.ReadUvarint(); err != nil {
		return
	}
	return
}
