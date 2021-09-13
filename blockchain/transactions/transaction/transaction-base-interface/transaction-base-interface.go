package transaction_base_interface

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/helpers"
)

type TransactionBaseInterface interface {
	helpers.SerializableInterface
	SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool)
	IncludeTransaction(blockHeight uint64, accsCollection *accounts.AccountsCollection, toks *tokens.Tokens) error
	ComputeFees() uint64
	ComputeAllKeys(out map[string]bool)
	VerifySignatureManually(hashForSignature []byte) bool
	Validate() error
	VerifyBloomAll() error
}
