package api_common

import (
	"net/http"
	"net/url"
	"pandora-pay/network/websocks/connection"
)

type APIBlockchain struct {
	Height            uint64 `json:"height" msgpack:"height"`
	Hash              string `json:"hash" msgpack:"hash"`
	PrevHash          string `json:"prevHash" msgpack:"prevHash"`
	KernelHash        string `json:"kernelHash" msgpack:"kernelHash"`
	PrevKernelHash    string `json:"prevKernelHash" msgpack:"prevKernelHash"`
	Timestamp         uint64 `json:"timestamp" msgpack:"timestamp"`
	TransactionsCount uint64 `json:"transactions" msgpack:"transactions"`
	AccountsCount     uint64 `json:"accounts" msgpack:"accounts"`
	AssetsCount       uint64 `json:"assets" msgpack:"assets"`
	Target            string `json:"target" msgpack:"target"`
	TotalDifficulty   string `json:"totalDifficulty" msgpack:"totalDifficulty"`
}

func (api *APICommon) Blockchain(r *http.Request, args *struct{}, reply *APIBlockchain) error {
	x := api.localChain.Load()
	*reply = *x
	return nil
}

func (api *APICommon) Chain(r *http.Request, args *struct{}, reply *APIBlockchain) error {
	return api.Blockchain(r, args, reply)
}

func (api *APICommon) GetBlockchain_http(values url.Values) (interface{}, error) {
	reply := &APIBlockchain{}
	return reply, api.Blockchain(nil, nil, reply)
}

func (api *APICommon) GetBlockchain_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &APIBlockchain{}
	return reply, api.Blockchain(nil, nil, reply)
}
