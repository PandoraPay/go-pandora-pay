package transaction_zether_payload_extra

import (
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraInterface interface {
	BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) error
	IncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) error
	Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) error
	Serialize(w *helpers.BufferWriter, inclSignature bool)
	Deserialize(r *helpers.BufferReader) error
	VerifyExtraSignature(hashForSignature []byte) bool
	ComputeAllKeys(out map[string]bool)
	UpdateStatement(payloadStatement *crypto.Statement) error
}

func SerializeToBytes(self TransactionZetherPayloadExtraInterface, inclSignature bool) []byte {
	w := helpers.NewBufferWriter()
	self.Serialize(w, inclSignature)
	return w.Bytes()
}
