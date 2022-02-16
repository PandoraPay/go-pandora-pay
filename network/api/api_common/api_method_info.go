package api_common

import (
	"net/http"
	"pandora-pay/config"
)

type APIInfoReply struct {
	Name       string `json:"name" msgpack:"name"`
	Version    string `json:"version" msgpack:"version"`
	Network    uint64 `json:"network" msgpack:"network"`
	CPUThreads int    `json:"CPUThreads" msgpack:"CPUThreads"`
}

func (api *APICommon) GetInfo(r *http.Request, args *struct{}, reply *APIInfoReply) error {
	reply.Name = config.NAME
	reply.Version = config.VERSION
	reply.Network = config.NETWORK_SELECTED
	reply.CPUThreads = config.CPU_THREADS
	return nil
}
