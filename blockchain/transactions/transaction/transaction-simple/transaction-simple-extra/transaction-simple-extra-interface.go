package transaction_simple_extra

import (
	plain_account "pandora-pay/blockchain/data/plain-accounts/plain-account"
	"pandora-pay/helpers"
)

type TransactionSimpleExtraInterface interface {
	IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount) error
	Serialize(w *helpers.BufferWriter)
	Deserialize(r *helpers.BufferReader) error
	Validate() error
}
