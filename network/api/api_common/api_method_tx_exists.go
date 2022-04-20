package api_common

import (
	"net/http"
	"pandora-pay/helpers"
)

type APITxExistsRequest struct {
	Hash helpers.Base64 `json:"hash"  msgpack:"hash"`
}

type APITxExistsReply struct {
	Exists bool `json:"exists" msgpack:"exists"`
}

func (api *APICommon) GetTxExists(r *http.Request, args *APITxExistsRequest, reply *APITxExistsReply) (err error) {
	reply.Exists, err = api.ApiStore.chain.OpenExistsTx(args.Hash)
	return
}
