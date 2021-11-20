package api_common

import (
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
	"strconv"
)

type APITxHeight uint64

func (api *APICommon) GetTxHash(txHeight uint64) (helpers.HexBytes, error) {
	return api.ApiStore.openLoadTxHash(txHeight)
}

func (api *APICommon) GetTxHash_http(values *url.Values) (interface{}, error) {

	if values.Get("height") != "" {
		height, err := strconv.ParseUint(values.Get("height"), 10, 64)
		if err != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}
		return api.getBlockHash(height)
	}

	return nil, errors.New("parameter `height` is missing")
}

func (api *APICommon) GetTxHash_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := APITxHeight(0)
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.getBlockHash(uint64(request))
}
