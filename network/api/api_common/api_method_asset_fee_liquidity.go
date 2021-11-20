package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIAssetFeeLiquidityFeeRequest struct {
	api_types.APIHeightHash
}

type APIAssetFeeLiquidity struct {
	Asset        helpers.HexBytes `json:"asset"`
	Rate         uint64           `json:"rate"`
	LeadingZeros byte             `json:"leadingZeros"`
	Collector    helpers.HexBytes `json:"collector"` //collector Public Key
}

func (api *APICommon) getAssetFeeLiquidity(request *APIAssetFeeLiquidityFeeRequest) ([]byte, error) {
	out, err := api.ApiStore.openLoadAssetFeeLiquidity(request.Hash, request.Height)
	if err != nil || out == nil {
		return nil, err
	}
	return json.Marshal(out)
}

func (api *APICommon) GetAssetFeeLiquidity_http(values *url.Values) (interface{}, error) {
	request := &APIAssetFeeLiquidityFeeRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}
	return api.getAssetFeeLiquidity(request)
}

func (api *APICommon) GetAssetFeeLiquidity_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIAssetFeeLiquidityFeeRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.getAssetFeeLiquidity(request)
}
