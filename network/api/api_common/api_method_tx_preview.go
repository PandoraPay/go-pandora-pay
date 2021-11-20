package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/blockchain/info"
	"pandora-pay/cryptography"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APITransactionPreviewRequest struct {
	api_types.APIHeightHash
}

type APITransactionPreviewAnswer struct {
	TxPreview *info.TxPreview `json:"txPreview,omitempty"`
	Mempool   bool            `json:"mempool,omitempty"`
	Info      *info.TxInfo    `json:"info,omitempty"`
}

func (api *APICommon) GetTxPreview(request *APITransactionPreviewRequest) ([]byte, error) {
	var txPreview *info.TxPreview
	var txInfo *info.TxInfo
	var err error

	mempool := false
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		txMemPool := api.mempool.Txs.Get(string(request.Hash))
		if txMemPool != nil {
			mempool = true
			if txPreview, err = info.CreateTxPreviewFromTx(txMemPool.Tx); err != nil {
				return nil, err
			}
		} else {
			txPreview, txInfo, err = api.ApiStore.openLoadTxPreview(request.Hash, 0)
		}
	} else {
		txPreview, txInfo, err = api.ApiStore.openLoadTxPreview(nil, request.Height)
	}

	if err != nil || txPreview == nil {
		return nil, err
	}

	result := &APITransactionPreviewAnswer{txPreview, mempool, txInfo}
	return json.Marshal(result)
}

func (api *APICommon) GetTxPreview_http(values *url.Values) (interface{}, error) {

	request := &APITransactionPreviewRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.GetTxPreview(request)
}

func (api *APICommon) GetTxPreview_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APITransactionPreviewRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.GetTxPreview(request)
}
