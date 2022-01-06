package api_common

import (
	"encoding/json"
	"net/http"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
)

type APIBlockHashRequest struct {
	Height uint64 `json:"height"`
}

func (api *APICommon) BlockHash(r *http.Request, args *APIBlockHashRequest, reply *helpers.HexBytes) (err error) {
	*reply, err = api.ApiStore.chain.OpenLoadBlockHash(args.Height)
	return
}

func (api *APICommon) GetBlockHash_http(values url.Values) (interface{}, error) {
	args := &APIBlockHashRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	var reply helpers.HexBytes
	return reply, api.BlockHash(nil, args, &reply)
}

func (api *APICommon) GetBlockHash_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIBlockHashRequest{0}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	var reply helpers.HexBytes
	return reply, api.BlockHash(nil, args, &reply)
}
