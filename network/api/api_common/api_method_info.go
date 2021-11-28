package api_common

import (
	"net/http"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/network/websocks/connection"
)

type AppInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Network    uint64 `json:"network"`
	CPUThreads int    `json:"CPUThreads"`
}

func (api *APICommon) Info(r *http.Request, args *struct{}, reply *AppInfo) error {
	reply.Name = config.NAME
	reply.Version = config.VERSION
	reply.Network = config.NETWORK_SELECTED
	reply.CPUThreads = config.CPU_THREADS
	return nil
}

func (api *APICommon) GetInfo_http(values url.Values) (interface{}, error) {
	reply := &AppInfo{}
	return reply, api.Info(nil, &struct{}{}, reply)
}

func (api *APICommon) GetInfo_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &AppInfo{}
	return reply, api.Info(nil, &struct{}{}, reply)
}
