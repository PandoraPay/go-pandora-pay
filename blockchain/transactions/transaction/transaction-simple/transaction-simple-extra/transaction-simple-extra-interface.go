package transaction_simple_extra

import (
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/helpers"
)

type TransactionSimpleExtraInterface interface {
	IncludeTransactionVin0(blockHeight uint64, acc *account.Account) error
	Serialize(writer *helpers.BufferWriter)
	Deserialize(reader *helpers.BufferReader) error
	Validate() error
}
