package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIAccountMempoolRequest struct {
	api_types.APIAccountBaseRequest
}

func (api *APICommon) GetAccountMempool(request *APIAccountMempoolRequest) ([]byte, error) {

	publicKey, err := request.GetPublicKey()
	if err != nil {
		return nil, err
	}

	txs := api.mempool.Txs.GetAccountTxs(publicKey)

	var answer []helpers.HexBytes
	if txs != nil {
		answer = make([]helpers.HexBytes, len(txs))
		c := 0
		for _, tx := range txs {
			answer[c] = tx.Tx.Bloom.Hash
			c += 1
		}
	}

	return json.Marshal(answer)
}

func (api *APICommon) GetAccountMempool_http(values *url.Values) (interface{}, error) {

	request := &APIAccountMempoolRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.GetAccountMempool(request)
}

func (api *APICommon) GetAccountMempool_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIAccountMempoolRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.GetAccountMempool(request)
}
