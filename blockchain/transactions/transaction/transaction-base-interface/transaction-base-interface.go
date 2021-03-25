package transaction_base_interface

import (
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/helpers"
)

type TransactionBaseInterface interface {
	IncludeTransaction(blockHeight uint64, accs *accounts.Accounts, toks *tokens.Tokens) error
	ComputeFees(out map[string]uint64) (err error)
	VerifySignatureManually(hashForSignature []byte) bool
	Serialize(writer *helpers.BufferWriter, inclSignature bool)
	Validate() error
	Deserialize(reader *helpers.BufferReader) error
	VerifyBloomAll() error
}
