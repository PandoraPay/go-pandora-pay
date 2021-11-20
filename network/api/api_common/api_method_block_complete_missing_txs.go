package api_common

import (
	"encoding/json"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
)

type APIBlockCompleteMissingTxsRequest struct {
	Hash       helpers.HexBytes `json:"hash,omitempty"`
	MissingTxs []int            `json:"missingTxs,omitempty"`
}

type APIBlockCompleteMissingTxs struct {
	Txs []helpers.HexBytes `json:"txs,omitempty"`
}

func (api *APICommon) GetBlockCompleteMissingTxs(request *APIBlockCompleteMissingTxsRequest) ([]byte, error) {

	var blockCompleteMissingTxs *APIBlockCompleteMissingTxs
	var err error

	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockCompleteMissingTxs, err = api.ApiStore.openLoadBlockCompleteMissingTxsFromHash(request.Hash, request.MissingTxs)
	}
	if err != nil || blockCompleteMissingTxs == nil {
		return nil, err
	}
	return json.Marshal(blockCompleteMissingTxs)
}

func (api *APICommon) GetBlockCompleteMissingTxs_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &APIBlockCompleteMissingTxsRequest{nil, []int{}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return api.GetBlockCompleteMissingTxs(request)
}
