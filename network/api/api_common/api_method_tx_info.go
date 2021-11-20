package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APITransactionInfoRequest struct {
	api_types.APIHeightHash
}

func (api *APICommon) GetTxInfo(request *APITransactionInfoRequest) ([]byte, error) {
	txInfo, err := api.ApiStore.openLoadTxInfo(request.Hash, request.Height)
	if err != nil || txInfo == nil {
		return nil, err
	}
	return json.Marshal(txInfo)
}

func (api *APICommon) GetTxInfo_http(values *url.Values) (interface{}, error) {

	request := &APITransactionInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.GetTxInfo(request)
}

func (api *APICommon) GetTxInfo_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APITransactionInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.GetTxInfo(request)
}
