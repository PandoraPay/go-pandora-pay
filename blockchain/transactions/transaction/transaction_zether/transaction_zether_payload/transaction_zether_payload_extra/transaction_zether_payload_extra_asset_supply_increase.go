package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadExtraAssetSupplyIncrease struct {
	TransactionZetherPayloadExtraInterface
	AssetId           []byte
	ReceiverPublicKey []byte //must be registered before
	Value             uint64
	AssetSignature    []byte
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) BeforeIncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) IncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) Validate(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) error {
	if payloadExtra.Value == 0 {
		return errors.New("Asset Supply must be greater than zero")
	}
	if !bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("payloadAsset must be NATIVE_ASSET_FULL")
	}
	return nil
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	w.Write(payloadExtra.AssetId)
	w.Write(payloadExtra.ReceiverPublicKey)
	w.WriteUvarint(payloadExtra.Value)
	if inclSignature {
		w.Write(payloadExtra.AssetSignature)
	}
}

func (payloadExtra *TransactionZetherPayloadExtraAssetSupplyIncrease) Deserialize(r *helpers.BufferReader) (err error) {
	if payloadExtra.AssetId, err = r.ReadBytes(config_coins.ASSET_LENGTH); err != nil {
		return
	}
	if payloadExtra.ReceiverPublicKey, err = r.ReadBytes(cryptography.PublicKeySize); err != nil {
		return
	}
	if payloadExtra.Value, err = r.ReadUvarint(); err != nil {
		return
	}
	if payloadExtra.AssetSignature, err = r.ReadBytes(cryptography.SignatureSize); err != nil {
		return
	}
	return
}
