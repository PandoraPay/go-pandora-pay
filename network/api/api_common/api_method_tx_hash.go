package api_common

import (
	"encoding/json"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITxHashRequest struct {
	Height uint64 `json:"height"`
}

func (api *APICommon) TxHash(r *http.Request, args *APITxHashRequest, reply *helpers.HexBytes) (err error) {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		*reply, err = api.ApiStore.loadTxHash(reader, args.Height)
		return
	})
}

func (api *APICommon) GetTxHash_http(values url.Values) (interface{}, error) {
	args := &APITxHashRequest{0}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	var out helpers.HexBytes
	return out, api.TxHash(nil, args, &out)
}

func (api *APICommon) GetTxHash_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APITxHashRequest{0}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	var out helpers.HexBytes
	return out, api.TxHash(nil, args, &out)
}
