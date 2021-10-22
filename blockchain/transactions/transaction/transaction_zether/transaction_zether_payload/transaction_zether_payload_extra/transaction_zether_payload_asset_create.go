package transaction_zether_payload_extra

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_registrations"
	"pandora-pay/config/config_coins"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadAssetCreate struct {
	TransactionZetherPayloadExtraInterface
	AssetInfo *asset.Asset
}

func (payloadExtra *TransactionZetherPayloadAssetCreate) BeforeIncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex int, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyListByCounter [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadAssetCreate) IncludeTxPayload(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex int, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyListByCounter [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	list := helpers.NewBufferWriter()
	list.WriteUvarint(uint64(payloadIndex))
	list.WriteUvarint(blockHeight)

	hash := cryptography.SHA3(list.Bytes())
	if err = dataStorage.Asts.CreateAsset(hash, payloadExtra.AssetInfo); err != nil {
		return
	}

	return
}

func (payloadExtra *TransactionZetherPayloadAssetCreate) Validate(txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) error {

	if payloadExtra.AssetInfo.Supply != 0 {
		return errors.New("AssetInfo Supply must be zero")
	}
	if !bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("payloadAsset must be NATIVE_ASSET_FULL")
	}

	return payloadExtra.AssetInfo.Validate()
}

func (payloadExtra *TransactionZetherPayloadAssetCreate) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	payloadExtra.AssetInfo.Serialize(w)
}

func (payloadExtra *TransactionZetherPayloadAssetCreate) Deserialize(r *helpers.BufferReader) (err error) {
	return payloadExtra.AssetInfo.Deserialize(r)
}
