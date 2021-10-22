package transaction_zether_payload_extra

import (
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/helpers"
)

type TransactionZetherPayloadAssetCreate struct {
	TransactionZetherPayloadExtraInterface
	AssetInfo *asset.Asset
}

func (payloadExtra *TransactionZetherPayloadAssetCreate) Serialize(w *helpers.BufferWriter, inclSignature bool) {
	payloadExtra.AssetInfo.Serialize(w)
}

func (payloadExtra *TransactionZetherPayloadAssetCreate) Deserialize(r *helpers.BufferReader) (err error) {
	return payloadExtra.AssetInfo.Deserialize(r)
}
