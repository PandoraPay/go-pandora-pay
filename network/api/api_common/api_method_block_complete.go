package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/cryptography"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIBlockCompleteRequest struct {
	api_types.APIHeightHash
	ReturnType api_types.APIReturnType `json:"returnType,omitempty"`
}

func (api *APICommon) getBlockComplete(request *APIBlockCompleteRequest) ([]byte, error) {

	var blockComplete *block_complete.BlockComplete
	var err error

	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		blockComplete, err = api.ApiStore.openLoadBlockCompleteFromHash(request.Hash)
	} else {
		blockComplete, err = api.ApiStore.openLoadBlockCompleteFromHeight(request.Height)
	}
	if err != nil || blockComplete == nil {
		return nil, err
	}
	if request.ReturnType == api_types.RETURN_SERIALIZED {
		return blockComplete.BloomBlkComplete.Serialized, nil
	}
	return json.Marshal(blockComplete)
}

func (api *APICommon) GetBlockComplete_http(values *url.Values) (interface{}, error) {

	request := &APIBlockCompleteRequest{api_types.APIHeightHash{0, nil}, api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getBlockComplete(request)
}

func (api *APICommon) GetBlockComplete_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &APIBlockCompleteRequest{api_types.APIHeightHash{0, nil}, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return api.getBlockComplete(request)
}
