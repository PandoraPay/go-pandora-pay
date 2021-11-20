package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"strconv"
)

type APIAccountTxsRequest struct {
	api_types.APIAccountBaseRequest
	Next uint64 `json:"next,omitempty"`
}

type APIAccountTxsAnswer struct {
	Count uint64             `json:"count,omitempty"`
	Txs   []helpers.HexBytes `json:"txs,omitempty"`
}

func (api *APICommon) getAccountTxs(request *APIAccountTxsRequest) ([]byte, error) {

	publicKey, err := request.GetPublicKey()
	if err != nil {
		return nil, err
	}

	answer, err := api.ApiStore.openLoadAccountTxsFromPublicKey(publicKey, request.Next)
	if err != nil || answer == nil {
		return nil, err
	}

	return json.Marshal(answer)
}

func (api *APICommon) GetAccountTxs_http(values *url.Values) (interface{}, error) {

	request := &APIAccountTxsRequest{}

	var err error
	if values.Get("next") != "" {
		if request.Next, err = strconv.ParseUint(values.Get("next"), 10, 64); err != nil {
			return nil, err
		}
	}

	if err = request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getAccountTxs(request)
}

func (api *APICommon) GetAccountTxs_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIAccountTxsRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.getAccountTxs(request)
}
