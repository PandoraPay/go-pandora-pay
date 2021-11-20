package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/network/websocks/connection"
)

type appInfo struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Network    uint64 `json:"network"`
	CPUThreads int    `json:"CPUThreads"`
}

func (api *APICommon) getInfo() ([]byte, error) {
	return json.Marshal(&appInfo{
		Name:       config.NAME,
		Version:    config.VERSION,
		Network:    config.NETWORK_SELECTED,
		CPUThreads: config.CPU_THREADS,
	})
}

func (api *APICommon) GetInfo_http(values *url.Values) (interface{}, error) {
	return api.getInfo()
}

func (api *APICommon) GetInfo_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.getInfo()
}
