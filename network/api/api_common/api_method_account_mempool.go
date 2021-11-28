package api_common

import (
	"encoding/json"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIAccountMempoolRequest struct {
	api_types.APIAccountBaseRequest
}

type APIAccountMempoolReply struct {
	List []helpers.HexBytes `json:"list"`
}

func (api *APICommon) AccountMempool(r *http.Request, args *APIAccountMempoolRequest, reply *APIAccountMempoolReply) error {

	publicKey, err := args.GetPublicKey()
	if err != nil {
		return err
	}

	txs := api.mempool.Txs.GetAccountTxs(publicKey)

	if txs != nil {
		reply.List = make([]helpers.HexBytes, len(txs))
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
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIAccountMempoolReply{}
	return reply, api.AccountMempool(nil, args, reply)
}

func (api *APICommon) GetAccountMempool_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAccountMempoolRequest{}
	if err := json.Unmarshal(values, &args); err != nil {
		return nil, err
	}
	reply := &APIAccountMempoolReply{}
	return reply, api.AccountMempool(nil, args, reply)
}
