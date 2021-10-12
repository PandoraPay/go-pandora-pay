package transaction_zether_extra

import (
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

type TransactionZetherExtraInterface interface {
	IncludeTransaction(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, dataStorage *data_storage.DataStorage) error
	Serialize(w *helpers.BufferWriter)
	Deserialize(r *helpers.BufferReader) error
	Validate() error
}
