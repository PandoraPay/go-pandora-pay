package transaction_zether_payload_extra

import (
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraInterface interface {
	BeforeIncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyListByCounter [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) error
	IncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyListByCounter [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) error
	Validate(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) error
	Serialize(w *helpers.BufferWriter, inclSignature bool)
	Deserialize(r *helpers.BufferReader) error
	VerifyExtraSignature(hashForSignature []byte) bool
}
