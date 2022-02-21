package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraStaking struct {
	TransactionZetherPayloadExtraInterface
}

func (payloadExtra *TransactionZetherPayloadExtraStaking) BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) error {
	return nil
}

func (payloadExtra *TransactionZetherPayloadExtraStaking) AfterIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	return
}

func (payloadExtra *TransactionZetherPayloadExtraStaking) ComputeAllKeys(out map[string]bool) {
}

func (payloadExtra *TransactionZetherPayloadExtraStaking) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) (err error) {

	if bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) == false {
		return errors.New("Payload[0] asset must be a native asset")
	}

	if payloadBurnValue == 0 {
		return errors.New("Payload burn value must be greater than zero")
	}

	return
}

func (payloadExtra *TransactionZetherPayloadExtraStaking) Serialize(w *helpers.BufferWriter, inclSignature bool) {
}

func (payloadExtra *TransactionZetherPayloadExtraStaking) Deserialize(r *helpers.BufferReader) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraStaking) UpdateStatement(payloadStatement *crypto.Statement) error {
	return nil
}
