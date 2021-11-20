package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/blockchain/blocks/block"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIBlockRequest struct {
	api_types.APIHeightHash
	ReturnType api_types.APIReturnType `json:"returnType,omitempty"`
}

type APIBlockWithTxsAnswer struct {
	Block           *block.Block       `json:"block,omitempty"`
	BlockSerialized helpers.HexBytes   `json:"serialized,omitempty"`
	Txs             []helpers.HexBytes `json:"txs,omitempty"`
}

func (api *APICommon) getBlock(request *APIBlockRequest) ([]byte, error) {

	var out *APIBlockWithTxsAnswer

	var err error
	if request.Hash != nil && len(request.Hash) == cryptography.HashSize {
		out, err = api.ApiStore.openLoadBlockWithTXsFromHash(request.Hash)
	} else {
		out, err = api.ApiStore.openLoadBlockWithTXsFromHeight(request.Height)
	}
	if err != nil || out.Block == nil {
		return nil, err
	}

	if request.ReturnType == api_types.RETURN_SERIALIZED {
		out.BlockSerialized = helpers.SerializeToBytes(out.Block)
		out.Block = nil
	}

	return json.Marshal(out)
}

func (api *APICommon) GetBlock_http(values *url.Values) (interface{}, error) {

	request := &APIBlockRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getBlock(request)
}

func (api *APICommon) GetBlock_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &APIBlockRequest{api_types.APIHeightHash{0, nil}, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.getBlock(request)
}
