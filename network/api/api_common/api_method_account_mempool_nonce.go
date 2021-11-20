package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIAccountMempoolNonceRequest struct {
	api_types.APIAccountBaseRequest
}

func (api *APICommon) getAccountMempoolNonce(request *APIAccountMempoolNonceRequest) ([]byte, error) {
	publicKey, err := request.GetPublicKey()
	if err != nil {
		return nil, err
	}

	nonce, err := api.ApiStore.OpenLoadPlainAccountNonceFromPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	return json.Marshal(api.mempool.GetNonce(publicKey, nonce))
}

func (api *APICommon) GetAccountMempoolNonce_http(values *url.Values) (interface{}, error) {
	request := &APIAccountMempoolNonceRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getAccountMempoolNonce(request)
}

func (api *APICommon) GetAccountMempoolNonce_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIAccountMempoolNonceRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.getAccountMempoolNonce(request)
}
