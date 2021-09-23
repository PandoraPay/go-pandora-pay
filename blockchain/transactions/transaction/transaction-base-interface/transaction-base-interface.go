package transaction_base_interface

import (
	"pandora-pay/blockchain/data/accounts"
	plain_accounts "pandora-pay/blockchain/data/plain-accounts"
	"pandora-pay/blockchain/data/registrations"
	"pandora-pay/blockchain/data/tokens"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction-data"
	"pandora-pay/helpers"
)

type TransactionBaseInterface interface {
	helpers.SerializableInterface
	SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool)
	IncludeTransaction(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, regs *registrations.Registrations, plainAccs *plain_accounts.PlainAccounts, accsCollection *accounts.AccountsCollection, toks *tokens.Tokens) error
	ComputeFees() (uint64, error)
	ComputeAllKeys(out map[string]bool)
	VerifySignatureManually(hashForSignature []byte) bool
	Validate() error
	VerifyBloomAll() error
}
