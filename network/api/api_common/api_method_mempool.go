package api_common

import (
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
)

type APIMempoolRequest struct {
	ChainHash helpers.HexBytes `json:"chainHash,omitempty" msgpack:"chainHash,omitempty"`
	Page      int              `json:"page,omitempty" msgpack:"page,omitempty"`
	Count     int              `json:"count,omitempty" msgpack:"count,omitempty"`
}

type APIMempoolReply struct {
	ChainHash helpers.HexBytes   `json:"chainHash" msgpack:"chainHash"`
	Count     int                `json:"count" msgpack:"count"`
	Hashes    []helpers.HexBytes `json:"hashes" msgpack:"hashes"`
}

func (api *APICommon) Mempool(r *http.Request, args *APIMempoolRequest, reply *APIMempoolReply) error {

	transactions, finalChainHash := api.mempool.GetNextTransactionsToInclude(args.ChainHash)

	if args.Count == 0 {
		args.Count = config.API_MEMPOOL_MAX_TRANSACTIONS
	}

	start := generics.Max(args.Page*args.Count, 0)

	length := generics.Min(generics.Max(len(transactions)-start, 0), config.API_MEMPOOL_MAX_TRANSACTIONS)

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
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIMempoolReply{}
	return reply, api.Mempool(nil, args, reply)
}

func (api *APICommon) GetMempool_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIMempoolRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIMempoolReply{}
	return reply, api.Mempool(nil, args, reply)
}
