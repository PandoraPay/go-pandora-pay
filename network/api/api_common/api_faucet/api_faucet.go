package api_faucet

import (
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/network/websocks/connection"
)

func (api *APICommonFaucet) GetFaucetInfoHttp(values *url.Values) (interface{}, error) {
	return api.GetFaucetInfo()
}

func (api *APICommonFaucet) GetFaucetInfoWebsocket(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.GetFaucetInfo()
}

func (api *APICommonFaucet) GetFaucetCoinsHttp(values *url.Values) (interface{}, error) {

	request := &APIFaucetCoinsRequest{"", ""}

	if values.Get("address") != "" {
		request.Address = values.Get("address")
	} else {
		return nil, errors.New("parameter 'address' was not specified")
	}

	if values.Get("faucetToken") != "" {
		request.FaucetToken = values.Get("faucetToken")
	}

	return api.GetFaucetCoins(request)
}

func (api *APICommonFaucet) GetFaucetCoinsWebsocket(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIFaucetCoinsRequest{"", ""}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}
	return api.GetFaucetCoins(request)
}
