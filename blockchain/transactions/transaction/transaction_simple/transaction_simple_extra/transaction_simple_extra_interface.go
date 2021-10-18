package transaction_simple_extra

import (
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/helpers"
)

type TransactionSimpleExtraInterface interface {
	IncludeTransactionVin0(blockHeight uint64, plainAcc *plain_account.PlainAccount, dataStorage *data_storage.DataStorage) error
	Serialize(w *helpers.BufferWriter, inclSignature bool)
	Deserialize(r *helpers.BufferReader) error
	Validate() error
}
