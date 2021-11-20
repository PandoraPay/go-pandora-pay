package api_common

import (
	"encoding/hex"
	"encoding/json"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
	"strconv"
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

func (api *APICommon) GetMempool(request *APIMempoolRequest) ([]byte, error) {

	transactions, finalChainHash := api.mempool.GetNextTransactionsToInclude(request.ChainHash)

	if request.Count == 0 {
		request.Count = config.API_MEMPOOL_MAX_TRANSACTIONS
	}

	start := request.Page * request.Count

	length := len(transactions) - start
	if length < 0 {
		length = 0
	}
	if length > config.API_MEMPOOL_MAX_TRANSACTIONS {
		length = config.API_MEMPOOL_MAX_TRANSACTIONS
	}

	result := &APIMempoolAnswer{
		Count:  len(transactions),
		Hashes: make([]helpers.HexBytes, length),
	}

	if request.ChainHash == nil {
		result.ChainHash = finalChainHash
	}

	for i := range result.Hashes {
		result.Hashes[i] = transactions[start+i].Bloom.Hash
	}

	return json.Marshal(result)
}

func (api *APICommon) GetMempool_http(values *url.Values) (interface{}, error) {
	request := &APIMempoolRequest{}

	var err error
	if values.Get("chainHash") != "" {
		if request.ChainHash, err = hex.DecodeString(values.Get("chainHash")); err != nil {
			return nil, err
		}
	}
	if values.Get("page") != "" {
		if request.Page, err = strconv.Atoi(values.Get("page")); err != nil {
			return nil, err
		}
	}
	if values.Get("count") != "" {
		if request.Count, err = strconv.Atoi(values.Get("count")); err != nil {
			return nil, err
		}
	}

	return api.GetMempool(request)
}

func (api *APICommon) GetMempool_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIMempoolRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.GetMempool(request)
}
