package api_delegator_node

import (
	"encoding/json"
	"net/url"
	"pandora-pay/config/config_nodes"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
)

type ApiDelegatorNodeInfoRequest struct {
}

type ApiDelegatorNodeInfoAnswer struct {
	MaximumAllowed int              `json:"maximumAllowed"`
	DelegatesCount int              `json:"delegatesCount"`
	DelegatesFee   uint64           `json:"delegatesFee"`
	Challenge      helpers.HexBytes `json:"challenge"`
}

func (api *DelegatorNode) getDelegatorNodeInfo(request *ApiDelegatorNodeInfoRequest) ([]byte, error) {

	answer := &ApiDelegatorNodeInfoAnswer{
		config_nodes.DELEGATES_MAXIMUM,
		api.wallet.GetDelegatesCount(),
		config_nodes.DELEGATOR_FEE,
		api.challenge,
	}

	return json.Marshal(answer)
}

func (api *DelegatorNode) GetDelegatorNodeInfo_http(values *url.Values) (interface{}, error) {
	request := &ApiDelegatorNodeInfoRequest{}
	return api.getDelegatorNodeInfo(request)
}

func (api *DelegatorNode) GetDelegatorNodeInfo_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &ApiDelegatorNodeInfoRequest{}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.getDelegatorNodeInfo(request)
}
