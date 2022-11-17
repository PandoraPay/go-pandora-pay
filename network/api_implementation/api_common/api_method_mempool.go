package api_common

import (
	"net/http"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
)

type APIMempoolRequest struct {
	ChainHash helpers.Base64 `json:"chainHash,omitempty" msgpack:"chainHash,omitempty"`
	Page      int            `json:"page,omitempty" msgpack:"page,omitempty"`
	Count     int            `json:"count,omitempty" msgpack:"count,omitempty"`
}

type APIMempoolReply struct {
	ChainHash []byte   `json:"chainHash" msgpack:"chainHash"`
	Count     int      `json:"count" msgpack:"count"`
	Hashes    [][]byte `json:"hashes" msgpack:"hashes"`
}

func (api *APICommon) GetMempool(r *http.Request, args *APIMempoolRequest, reply *APIMempoolReply) error {

	transactions, finalChainHash := api.mempool.GetNextTransactionsToInclude(args.ChainHash)

	if args.Count == 0 {
		args.Count = config.API_MEMPOOL_MAX_TRANSACTIONS
	}

	start := generics.Max(args.Page*args.Count, 0)

	length := generics.Min(generics.Max(len(transactions)-start, 0), config.API_MEMPOOL_MAX_TRANSACTIONS)

	reply.Count = len(transactions)
	reply.Hashes = make([][]byte, length)

	if args.ChainHash == nil {
		reply.ChainHash = finalChainHash
	}

	for i := range reply.Hashes {
		reply.Hashes[i] = transactions[start+i].Bloom.Hash
	}

	return nil
}
