package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/network/websocks/connection"
)

func (api *APICommon) getPing() ([]byte, error) {
	return json.Marshal(&struct {
		Ping string `json:"ping"`
	}{Ping: "pong"})
}

func (api *APICommon) GetPing_http(values *url.Values) (interface{}, error) {
	return api.getPing()
}

func (api *APICommon) GetPing_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.getPing()
}
