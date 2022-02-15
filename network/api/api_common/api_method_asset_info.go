package api_common

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAssetInfoRequest struct {
	Height uint64 `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash   []byte `json:"hash,omitempty" msgpack:"hash,omitempty"`
}

func (api *APICommon) getAssetInfo(r *http.Request, args *APIAssetInfoRequest, reply *info.AssetInfo) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(args.Hash) == 0 {
			if args.Hash, err = api.ApiStore.loadAssetHash(reader, args.Height); err != nil {
				return
			}
		}

		data := reader.Get("assetInfo_ByHash:" + string(args.Hash))
		if data == nil {
			return errors.New("AssetInfo was not found")
		}

		return msgpack.Unmarshal(data, reply)
	})
}

func (api *APICommon) GetAssetInfo_http(values url.Values) (interface{}, error) {
	args := &APIAssetInfoRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &info.AssetInfo{}
	return reply, api.getAssetInfo(nil, args, reply)
}

func (api *APICommon) GetAssetInfo_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAssetInfoRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &info.AssetInfo{}
	return reply, api.getAssetInfo(nil, args, reply)
}
