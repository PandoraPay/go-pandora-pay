package transaction_base_interface

import (
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/helpers"
	"pandora-pay/helpers/advanced_buffers"
)

type TransactionBaseInterface interface {
	helpers.SerializableInterface
	SerializeAdvanced(w *advanced_buffers.BufferWriter, inclSignature bool)
	IncludeTransaction(blockHeight uint64, txHash []byte, dataStorage *data_storage.DataStorage) error
	ComputeFee() (uint64, error)
	ComputeAllKeys(out map[string]bool)
	VerifySignatureManually(hashForSignature []byte) bool
	Validate() error
	VerifyBloomAll() error
	BloomNow() error
	GetBloomExtra() any
	SetBloomExtra(bloom any)
}
