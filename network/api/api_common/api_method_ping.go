package api_common

import (
	"net/http"
	"net/url"
	"pandora-pay/network/websocks/connection"
)

type Pong struct {
	Ping string `json:"ping" msgpack:"ping"`
}

func (api *APICommon) Ping(r *http.Request, args *struct{}, reply *Pong) error {
	reply.Ping = "pong"
	return nil
}

func (api *APICommon) GetPing_http(values url.Values) (interface{}, error) {
	reply := &Pong{}
	return reply, api.Ping(nil, &struct{}{}, reply)
}

func (api *APICommon) GetPing_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &Pong{}
	return reply, api.Ping(nil, &struct{}{}, reply)
}
