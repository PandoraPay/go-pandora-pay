package transaction_base_interface

import (
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/helpers"
)

type TransactionBaseInterface interface {
	helpers.SerializableInterface
	ComputeExtraSpace() uint64
	SerializeAdvanced(w *helpers.BufferWriter, inclSignature bool)
	IncludeTransaction(blockHeight uint64, dataStorage *data_storage.DataStorage) error
	ComputeFees() (uint64, error)
	ComputeAllKeys(out map[string]bool)
	VerifySignatureManually(hashForSignature []byte) bool
	Validate() error
	VerifyBloomAll() error
}
