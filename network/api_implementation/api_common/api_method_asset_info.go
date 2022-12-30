package api_common

import (
	"errors"
	"net/http"
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers"
	"pandora-pay/helpers/msgpack"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAssetInfoRequest struct {
	Height uint64         `json:"height,omitempty" msgpack:"height,omitempty"`
	Hash   helpers.Base64 `json:"hash,omitempty" msgpack:"hash,omitempty"`
}

func (api *APICommon) GetAssetInfo(r *http.Request, args *APIAssetInfoRequest, reply *info.AssetInfo) error {
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
