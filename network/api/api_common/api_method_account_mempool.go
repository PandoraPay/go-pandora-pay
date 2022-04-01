package api_common

import (
	"net/http"
	"pandora-pay/network/api/api_common/api_types"
)

type APIAccountMempoolRequest struct {
	api_types.APIAccountBaseRequest
}

type APIAccountMempoolReply struct {
	List [][]byte `json:"list" msgpack:"list"`
}

func (api *APICommon) GetAccountMempool(r *http.Request, args *APIAccountMempoolRequest, reply *APIAccountMempoolReply) error {

	publicKeyHash, err := args.GetPublicKeyHash(true)
	if err != nil {
		return err
	}

	txs := api.mempool.Txs.GetAccountTxs(publicKeyHash)

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
