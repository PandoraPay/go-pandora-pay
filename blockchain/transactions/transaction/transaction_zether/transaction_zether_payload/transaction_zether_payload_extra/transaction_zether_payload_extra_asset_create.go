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
	"pandora-pay/helpers/advanced_buffers"
)

type TransactionZetherPayloadExtraAssetCreate struct {
	TransactionZetherPayloadExtraInterface
	Asset *asset.Asset
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) BeforeIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {
	return
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) GetAssetId(txHash []byte, payloadIndex byte) []byte {
	list := advanced_buffers.NewBufferWriter()
	list.WriteByte(payloadIndex)
	list.Write(txHash)
	return cryptography.RIPEMD(cryptography.SHA3(list.Bytes()))
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) AfterIncludeTxPayload(txHash []byte, payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, publicKeyList [][]byte, blockHeight uint64, dataStorage *data_storage.DataStorage) (err error) {

	hash := payloadExtra.GetAssetId(txHash, payloadIndex)

	if bytes.Equal(hash, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("invalid hash")
	}

	//existence verification is done in CreateAsset
	if err = dataStorage.Asts.CreateAsset(hash, payloadExtra.Asset); err != nil {
		return
	}

	return
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) Validate(payloadRegistrations *transaction_zether_registrations.TransactionZetherDataRegistrations, payloadIndex byte, payloadAsset []byte, payloadBurnValue uint64, payloadStatement *crypto.Statement, payloadParity bool) error {

	if payloadExtra.Asset.Supply != 0 {
		return errors.New("AssetInfo Supply must be zero")
	}
	if !bytes.Equal(payloadAsset, config_coins.NATIVE_ASSET_FULL) {
		return errors.New("payloadAsset must be NATIVE_ASSET_FULL")
	}

	return payloadExtra.Asset.Validate()
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) ComputeAllKeys(out map[string]bool) {

}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) Serialize(w *advanced_buffers.BufferWriter, inclSignature bool) {
	payloadExtra.Asset.Serialize(w)
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) Deserialize(r *advanced_buffers.BufferReader) (err error) {
	payloadExtra.Asset = asset.NewAsset(helpers.EmptyBytes(cryptography.PublicKeyHashSize), 0)
	return payloadExtra.Asset.Deserialize(r)
}

func (payloadExtra *TransactionZetherPayloadExtraAssetCreate) UpdateStatement(payloadStatement *crypto.Statement) error {
	return nil
}
