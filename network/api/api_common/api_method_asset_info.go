package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIAssetInfoRequest struct {
	api_types.APIHeightHash
}

func (api *APICommon) getAssetInfo(request *APIAssetInfoRequest) ([]byte, error) {
	astInfo, err := api.ApiStore.openLoadAssetInfo(request.Hash, request.Height)
	if err != nil || astInfo == nil {
		return nil, err
	}
	return json.Marshal(astInfo)
}

func (api *APICommon) GetAssetInfo_http(values *url.Values) (interface{}, error) {

	request := &APIAssetInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getAssetInfo(request)
}

func (api *APICommon) GetAssetInfo_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIAssetInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.getAssetInfo(request)
}
