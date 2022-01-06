package api_common

import (
	"encoding/json"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APITransactionInfoRequest struct {
	Height uint64           `json:"height,omitempty"`
	Hash   helpers.HexBytes `json:"hash,omitempty"`
}

func (api *APICommon) GetTxInfo(r *http.Request, args *APITransactionInfoRequest, reply *info.TxInfo) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(args.Hash) == 0 {
			if args.Hash, err = api.ApiStore.loadTxHash(reader, args.Height); err != nil {
				return
			}
		}

		return api.ApiStore.loadTxInfo(reader, args.Hash, reply)
	})
}

func (api *APICommon) GetTxInfo_http(values url.Values) (interface{}, error) {
	args := &APITransactionInfoRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &info.TxInfo{}
	return reply, api.GetTxInfo(nil, args, reply)
}

func (api *APICommon) GetTxInfo_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APITransactionInfoRequest{}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &info.TxInfo{}
	return reply, api.GetTxInfo(nil, args, reply)
}
