package transaction_simple_extra

import (
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_parts"
	"pandora-pay/helpers"
)

type TransactionSimpleExtraInterface interface {
	IncludeTransactionExtra(blockHeight uint64, vinPublicKeyHashes [][]byte, vin []*transaction_simple_parts.TransactionSimpleInput, vout []*transaction_simple_parts.TransactionSimpleOutput, dataStorage *data_storage.DataStorage) error
	Serialize(w *helpers.BufferWriter, vin []*transaction_simple_parts.TransactionSimpleInput, vout []*transaction_simple_parts.TransactionSimpleOutput, inclSignature bool)
	Deserialize(r *helpers.BufferReader, vin []*transaction_simple_parts.TransactionSimpleInput, vout []*transaction_simple_parts.TransactionSimpleOutput) error
	Validate(vin []*transaction_simple_parts.TransactionSimpleInput, vout []*transaction_simple_parts.TransactionSimpleOutput) error
}
