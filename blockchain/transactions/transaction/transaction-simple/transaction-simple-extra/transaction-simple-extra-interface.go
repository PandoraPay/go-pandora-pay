package transaction_simple_extra

import (
	"pandora-pay/blockchain/data/accounts"
	plain_accounts "pandora-pay/blockchain/data/plain-accounts"
	plain_account "pandora-pay/blockchain/data/plain-accounts/plain-account"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/data/tokens"
	"pandora-pay/helpers"
)

type TransactionSimpleExtraInterface interface {
	IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, regs *registrations.Registrations, plainAccs *plain_accounts.PlainAccounts, accsCollection *accounts.AccountsCollection, toks *tokens.Tokens) error
	Serialize(w *helpers.BufferWriter)
	Deserialize(r *helpers.BufferReader) error
	Validate() error
}
