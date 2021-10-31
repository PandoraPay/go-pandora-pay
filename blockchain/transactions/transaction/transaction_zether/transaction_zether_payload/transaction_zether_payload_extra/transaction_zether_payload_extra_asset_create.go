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

type TransactionZetherPayloadExtraAssetCreate struct {
	TransactionZetherPayloadExtraInterface
	Asset *asset.Asset
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) BeforeIncludeTxPayload(txHash []byte, txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) GetAssetId(txHash []byte, payloadIndex byte) []byte {
	list := helpers.NewBufferWriter()
	list.WriteByte(payloadIndex)
	list.Write(txHash)
	return cryptography.RIPEMD(cryptography.SHA3(list.Bytes()))
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) IncludeTxPayload(txHash []byte, txRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	var exists bool
	hash := payloadExtra.GetAssetId(txHash, payloadIndex)

	if exists, err = dataStorage.Asts.Exists(string(hash)); err != nil {
		return
	}
	if exists {
		return errors.New("Asset with this Id already exists")
	}
	if err = dataStorage.Asts.CreateAsset(hash, payloadExtra.Asset); err != nil {
		return
	}

	return
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement) error {

	if payloadExtra.Asset.Supply != 0 {
		return errors.New("AssetInfo Supply must be zero")
	}
	if !bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("payloadAsset must be NATIVE_ASSET_FULL")
	}

	return payloadExtra.Asset.Validate()
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	payloadExtra.Asset.Serialize(w)
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) Deserialize(r *helpers.BufferReader) (err error) {
	payloadExtra.Asset = &asset.Asset{}
	return payloadExtra.Asset.Deserialize(r)
}
