package api_common

import (
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIAccountMempoolRequest struct {
	api_types.APIAccountBaseRequest
}

type APIAccountMempoolReply struct {
	List [][]byte `json:"list" msgpack:"list"`
}

func (api *APICommon) AccountMempool(r *http.Request, args *APIAccountMempoolRequest, reply *APIAccountMempoolReply) error {

	publicKey, err := args.GetPublicKey(true)
	if err != nil {
		return err
	}

	txs := api.mempool.Txs.GetAccountTxs(publicKey)

	if txs != nil {
		reply.List = make([][]byte, len(txs))
		c := 0
		for _, tx := range txs {
			reply.List[c] = tx.Tx.Bloom.Hash
			c += 1
		}
	}

	return nil
}

func (api *APICommon) GetAccountMempool_http(values url.Values) (interface{}, error) {
	args := &APIAccountMempoolRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIAccountMempoolReply{}
	return reply, api.AccountMempool(nil, args, reply)
}

func (api *APICommon) GetAccountMempool_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAccountMempoolRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIAccountMempoolReply{}
	return reply, api.AccountMempool(nil, args, reply)
}
