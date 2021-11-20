package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIAssetRequest struct {
	api_types.APIHeightHash
	ReturnType api_types.APIReturnType `json:"returnType,omitempty"`
}

func (api *APICommon) getAsset(request *APIAssetRequest) ([]byte, error) {
	asset, err := api.ApiStore.openLoadAsset(request.Hash, request.Height)
	if err != nil || asset == nil {
		return nil, err
	}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return helpers.SerializeToBytes(asset), nil
	}
	return json.Marshal(asset)
}

func (api *APICommon) GetAsset_http(values *url.Values) (interface{}, error) {
	request := &APIAssetRequest{ReturnType: api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}
	return api.getAsset(request)
}

func (api *APICommon) GetAsset_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIAssetRequest{api_types.APIHeightHash{0, nil}, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.getAsset(request)
}
