package api_common

import (
	"errors"
	"net/http"
	"pandora-pay/cryptography"
)

type APIMempoolExistsRequest struct {
	Hash []byte `json:"hash" msgpack:"hash"`
}

type APIMempoolExistsReply struct {
	Result bool `json:"result" msgpack:"result"`
}

func (api *APICommon) GetMempoolExists(r *http.Request, args *APIMempoolExistsRequest, reply *APIMempoolExistsReply) error {
	if len(args.Hash) != cryptography.HashSize {
		return errors.New("TxId must be 32 byte")
	}
	reply.Result = api.mempool.Txs.Get(string(args.Hash)) != nil
	return nil
}
