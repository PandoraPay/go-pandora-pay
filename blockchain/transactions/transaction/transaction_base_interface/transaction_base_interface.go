package transaction_base_interface

import (
	"pandora-pay/blockchain/data_storage"
	transaction_data "pandora-pay/blockchain/transactions/transaction/transaction_data"
	"pandora-pay/helpers"
)

type TransactionBaseInterface interface {
	helpers.SerializableInterface
	SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool)
	IncludeTransaction(txRegistrations *transaction_data.TransactionDataTransactions, blockHeight uint64, dataStorage *data_storage.DataStorage) error
	ComputeFees() (uint64, error)
	ComputeAllKeys(out map[string]bool)
	VerifySignatureManually(hashForSignature []byte) bool
	Validate() error
	VerifyBloomAll() error
}
