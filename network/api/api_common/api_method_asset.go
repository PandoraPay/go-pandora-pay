package api_common

import (
	"net/http"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAssetRequest struct {
	Height     uint64                  `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash       helpers.Base64          `json:"hash,omitempty" msgpack:"hash,omitempty"`
	ReturnType api_types.APIReturnType `json:"returnType,omitempty" msgpack:"returnType,omitempty"`
}

type APIAssetReply struct {
	Asset      *asset.Asset `json:"asset,omitempty" msgpack:"asset,omitempty"`
	Serialized []byte       `json:"serialized,omitempty" msgpack:"serialized,omitempty"`
}

func (api *APICommon) GetAsset(r *http.Request, args *APIAssetRequest, reply *APIAssetReply) (err error) {
	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if args.Hash == nil {
			if args.Hash, err = api.ApiStore.loadAssetHash(reader, args.Height); err != nil {
				return
			}
		}

		reply.Asset, err = assets.NewAssets(reader).GetAsset(args.Hash)
		return
	}); err != nil || reply.Asset == nil {
		return helpers.ReturnErrorIfNot(err, "Asset was not found")
	}

	if args.ReturnType == api_types.RETURN_SERIALIZED {
		reply.Serialized = helpers.SerializeToBytes(reply.Asset)
		reply.Asset = nil
	}
	return
}
