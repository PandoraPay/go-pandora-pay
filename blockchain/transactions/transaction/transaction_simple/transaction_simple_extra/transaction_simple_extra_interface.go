package transaction_simple_extra

import (
	"pandora-pay/blockchain/data_storage"
	plain_account "pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

type TransactionSimpleExtraInterface interface {
	IncludeTransactionVin0(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) error
	Serialize(w *helpers.BufferWriter)
	Deserialize(r *helpers.BufferReader) error
	Validate() error
}
