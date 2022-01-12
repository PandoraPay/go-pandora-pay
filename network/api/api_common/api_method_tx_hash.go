package api_common

import (
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITxHashRequest struct {
	Height uint64 `json:"height" msgpack:"height"`
}

func (api *APICommon) TxHash(r *http.Request, args *APITxHashRequest, reply *helpers.HexBytes) (err error) {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		*reply, err = api.ApiStore.loadTxHash(reader, args.Height)
		return
	})
}

func (api *APICommon) GetTxHash_http(values url.Values) (interface{}, error) {
	args := &APITxHashRequest{0}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	var out helpers.HexBytes
	return out, api.TxHash(nil, args, &out)
}

func (api *APICommon) GetTxHash_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APITxHashRequest{0}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	var out helpers.HexBytes
	return out, api.TxHash(nil, args, &out)
}
