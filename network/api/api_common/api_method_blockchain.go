package api_common

import (
	"net/http"
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
	Supply            uint64 `json:"supply" msgpack:"supply"`
	TotalDifficulty   string `json:"totalDifficulty" msgpack:"totalDifficulty"`
}

func (api *APICommon) GetBlockchain(r *http.Request, args *struct{}, reply *APIBlockchain) error {
	x := api.localChain.Load()
	*reply = *x
	return nil
}
