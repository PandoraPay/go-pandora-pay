package transaction_simple_extra

import (
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
)

type TransactionSimpleExtraInterface interface {
	IncludeTransactionVin0(blockHeight uint64, acc *account.Account) error
	Serialize(w *helpers.BufferWriter)
	Deserialize(r *helpers.BufferReader) error
	Validate() error
}
