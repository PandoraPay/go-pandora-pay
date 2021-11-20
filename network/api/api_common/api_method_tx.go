package api_common

import (
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APITransactionRequest struct {
	api_types.APIHeightHash
	ReturnType api_types.APIReturnType `json:"returnType,omitempty"`
}

type APITransactionAnswer struct {
	Tx           *transaction.Transaction `json:"tx,omitempty"`
	TxSerialized helpers.HexBytes         `json:"serialized,omitempty"`
	Mempool      bool                     `json:"mempool,omitempty"`
	Info         *info.TxInfo             `json:"info,omitempty"`
}

func (api *APICommon) getTx(request *APITransactionRequest) ([]byte, error) {
	var tx *transaction.Transaction
	var err error

	mempool := false
	var txInfo *info.TxInfo
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		txMempool := api.mempool.Txs.Get(string(request.Hash))
		if txMempool != nil {
			mempool = true
			tx = txMempool.Tx
		} else {
			tx, txInfo, err = api.ApiStore.openLoadTx(request.Hash, 0)
		}
	} else {
		tx, txInfo, err = api.ApiStore.openLoadTx(nil, request.Height)
	}

	if err != nil || tx == nil {
		return nil, err
	}

	result := &APITransactionAnswer{nil, nil, mempool, txInfo}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		result.TxSerialized = tx.Bloom.Serialized
	} else if request.ReturnType == api_types.RETURN_JSON {
		result.Tx = tx
	} else {
		return nil, errors.New("Invalid return type")
	}

	return json.Marshal(result)
}

func (api *APICommon) GetTx_http(values *url.Values) (interface{}, error) {

	request := &APITransactionRequest{api_types.APIHeightHash{0, nil}, api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}

	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getTx(request)
}

func (api *APICommon) GetTx_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APITransactionRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return api.getTx(request)
}
