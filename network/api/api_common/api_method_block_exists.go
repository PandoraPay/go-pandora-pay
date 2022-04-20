package api_common

import (
	"net/http"
	"pandora-pay/helpers"
)

type APIBlockExistsRequest struct {
	Hash helpers.Base64 `json:"hash"  msgpack:"hash"`
}

type APIBlockExistsReply struct {
	Exists bool `json:"exists" msgpack:"exists"`
}

func (api *APICommon) GetBlockExists(r *http.Request, args *APIBlockExistsRequest, reply *APIBlockExistsReply) (err error) {
	reply.Exists, err = api.ApiStore.chain.OpenExistsBlock(args.Hash)
	return
}
