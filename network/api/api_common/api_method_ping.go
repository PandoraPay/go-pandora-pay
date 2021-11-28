package api_common

import (
	"net/url"
	"pandora-pay/network/websocks/connection"
)

type Pong struct {
	Ping string `json:"ping"`
}

func (api *APICommon) GetPing_rpc(args *struct{}, reply *Pong) error {
	reply.Ping = "pong"
	return nil
}

func (api *APICommon) GetPing_http(values url.Values) (interface{}, error) {
	reply := &Pong{}
	return reply, api.GetPing_rpc(&struct{}{}, reply)
}

func (api *APICommon) GetPing_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &Pong{}
	return reply, api.GetPing_rpc(&struct{}{}, reply)
}
