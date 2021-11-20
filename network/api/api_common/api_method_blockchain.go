package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/network/websocks/connection"
)

type APIBlockchain struct {
	Height            uint64 `json:"height"`
	Hash              string `json:"hash"`
	PrevHash          string `json:"prevHash"`
	KernelHash        string `json:"kernelHash"`
	PrevKernelHash    string `json:"prevKernelHash"`
	Timestamp         uint64 `json:"timestamp"`
	TransactionsCount uint64 `json:"transactions"`
	AccountsCount     uint64 `json:"accounts"`
	AssetsCount       uint64 `json:"assets"`
	Target            string `json:"target"`
	TotalDifficulty   string `json:"totalDifficulty"`
}

func (api *APICommon) getBlockchain() ([]byte, error) {
	return json.Marshal(api.localChain.Load().(*APIBlockchain))
}

func (api *APICommon) GetBlockchain_http(values *url.Values) (interface{}, error) {
	return api.getBlockchain()
}

func (api *APICommon) GetBlockchain_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.getBlockchain()
}
