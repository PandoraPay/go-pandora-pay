package api_common

import (
	"net/http"
)

type APIBlockHashRequest struct {
	Height uint64 `json:"height" msgpack:"height"`
}

type APIBlockHashReply struct {
	Hash []byte `json:"hash" msgpack:"hash"`
}

func (api *APICommon) GetBlockHash(r *http.Request, args *APIBlockHashRequest, reply *APIBlockHashReply) (err error) {
	reply.Hash, err = api.ApiStore.chain.OpenLoadBlockHash(args.Height)
	return
}
