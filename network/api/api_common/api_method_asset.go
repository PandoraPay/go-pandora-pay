package api_common

import (
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAssetRequest struct {
	Height     uint64                  `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash       helpers.Base64          `json:"hash,omitempty" msgpack:"hash,omitempty"`
	ReturnType api_types.APIReturnType `json:"returnType,omitempty" msgpack:"returnType,omitempty"`
}

type APIAssetReply struct {
	Asset      *asset.Asset `json:"account,omitempty" msgpack:"account,omitempty"`
	Serialized []byte       `json:"serialized,omitempty" msgpack:"serialized,omitempty"`
}

func (api *APICommon) Asset(r *http.Request, args *APIAssetRequest, reply *APIAssetReply) (err error) {
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

func (api *APICommon) GetAsset_http(values url.Values) (interface{}, error) {
	args := &APIAssetRequest{0, nil, api_types.RETURN_JSON}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIAssetReply{}
	return reply, api.Asset(nil, args, reply)
}

func (api *APICommon) GetAsset_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAssetRequest{0, nil, api_types.RETURN_SERIALIZED}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIAssetReply{}
	return reply, api.Asset(nil, args, reply)
}
