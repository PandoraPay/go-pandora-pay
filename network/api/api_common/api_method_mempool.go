package api_common

import (
	"encoding/json"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
)

type APIMempoolRequest struct {
	ChainHash helpers.HexBytes `json:"chainHash,omitempty"`
	Page      int              `json:"page,omitempty"`
	Count     int              `json:"count,omitempty"`
}

type APIMempoolAnswer struct {
	ChainHash helpers.HexBytes   `json:"chainHash"`
	Count     int                `json:"count"`
	Hashes    []helpers.HexBytes `json:"hashes"`
}

func (api *APICommon) Mempool(r *http.Request, args *APIMempoolRequest, reply *APIMempoolAnswer) error {

	transactions, finalChainHash := api.mempool.GetNextTransactionsToInclude(args.ChainHash)

	if args.Count == 0 {
		args.Count = config.API_MEMPOOL_MAX_TRANSACTIONS
	}

	start := args.Page * args.Count

	length := len(transactions) - start
	if length < 0 {
		length = 0
	}
	if length > config.API_MEMPOOL_MAX_TRANSACTIONS {
		length = config.API_MEMPOOL_MAX_TRANSACTIONS
	}

	reply.Count = len(transactions)
	reply.Hashes = make([]helpers.HexBytes, length)

	if args.ChainHash == nil {
		reply.ChainHash = finalChainHash
	}

	for i := range reply.Hashes {
		reply.Hashes[i] = transactions[start+i].Bloom.Hash
	}

	return nil
}

func (api *APICommon) GetMempool_http(values url.Values) (interface{}, error) {
	args := &APIMempoolRequest{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIMempoolAnswer{}
	return reply, api.Mempool(nil, args, reply)
}

func (api *APICommon) GetMempool_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIMempoolRequest{}
	if err := json.Unmarshal(values, &args); err != nil {
		return nil, err
	}
	reply := &APIMempoolAnswer{}
	return reply, api.Mempool(nil, args, reply)
}
