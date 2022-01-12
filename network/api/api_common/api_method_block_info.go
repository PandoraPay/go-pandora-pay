package api_common

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIBlockInfoRequest struct {
	Height uint64           `json:"height,omitempty"  msgpack:"height,omitempty"`
	Hash   helpers.HexBytes `json:"hash,omitempty"  msgpack:"hash,omitempty"`
}

func (api *APICommon) BlockInfo(r *http.Request, args *APIBlockInfoRequest, reply *info.BlockInfo) error {
	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		if len(args.Hash) == 0 {
			if args.Hash, err = api.ApiStore.chain.LoadBlockHash(reader, args.Height); err != nil {
				return
			}
		}

		data := reader.Get("blockInfo_ByHash" + string(args.Hash))
		if data == nil {
			return errors.New("BlockInfo was not found")
		}
		return msgpack.Unmarshal(data, reply)
	})
}

func (api *APICommon) GetBlockInfo_http(values url.Values) (interface{}, error) {
	args := &APIBlockInfoRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &info.BlockInfo{}
	return reply, api.BlockInfo(nil, args, reply)
}

func (api *APICommon) GetBlockInfo_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIBlockInfoRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &info.BlockInfo{}
	return reply, api.BlockInfo(nil, args, reply)
}
