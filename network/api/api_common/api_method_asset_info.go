package api_common

import (
	"encoding/json"
	"errors"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAssetInfoRequest struct {
	Height uint64           `json:"height,omitempty"`
	Hash   helpers.HexBytes `json:"hash,omitempty"`
}

func (api *APICommon) AssetInfo(r *http.Request, args *APIAssetInfoRequest, reply *info.AssetInfo) error {
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

		return json.Unmarshal(data, reply)
	})
}

func (api *APICommon) GetAssetInfo_http(values url.Values) (interface{}, error) {
	args := &APIAssetInfoRequest{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &info.AssetInfo{}
	return reply, api.AssetInfo(nil, args, reply)
}

func (api *APICommon) GetAssetInfo_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAssetInfoRequest{}
	if err := json.Unmarshal(values, &args); err != nil {
		return nil, err
	}
	reply := &info.AssetInfo{}
	return reply, api.AssetInfo(nil, args, reply)
}
