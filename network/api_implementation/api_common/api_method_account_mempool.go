package api_common

import (
	"net/http"
	"pandora-pay/network/api_implementation/api_common/api_types"
)

type APIAccountMempoolRequest struct {
	api_types.APIAccountBaseRequest
}

type APIAccountMempoolReply struct {
	List [][]byte `json:"list" msgpack:"list"`
}

func (api *APICommon) GetAccountMempool(r *http.Request, args *APIAccountMempoolRequest, reply *APIAccountMempoolReply) error {

	publicKey, err := args.GetPublicKey(true)
	if err != nil {
		return err
	}

	txs := api.mempool.Txs.GetAccountTxs(publicKey)

	if txs != nil {
		reply.List = make([][]byte, len(txs))
		c := 0
		for _, tx := range txs {
			reply.List[c] = tx.Tx.Bloom.Hash
			c += 1
		}
	}

	return nil
}
