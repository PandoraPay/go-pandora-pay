package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIBlockInfoRequest struct {
	api_types.APIHeightHash
}

func (api *APICommon) getBlockInfo(request *APIBlockInfoRequest) ([]byte, error) {
	blockInfo, err := api.ApiStore.openLoadBlockInfo(request.Height, request.Hash)
	if err != nil || blockInfo == nil {
		return nil, err
	}
	return json.Marshal(blockInfo)
}

func (api *APICommon) GetBlockInfo_http(values *url.Values) (interface{}, error) {

	request := &APIBlockInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getBlockInfo(request)
}

func (api *APICommon) GetBlockInfo_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &APIBlockInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.getBlockInfo(request)
}
