package api_delegator_node

import (
	"net/http"
	"net/url"
	"pandora-pay/config/config_nodes"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
)

type ApiDelegatorNodeInfoReply struct {
	MaximumAllowed int              `json:"maximumAllowed"`
	DelegatesCount int              `json:"delegatesCount"`
	DelegatesFee   uint64           `json:"delegatesFee"`
	Challenge      helpers.HexBytes `json:"challenge"`
}

func (api *DelegatorNode) DelegatorNodeInfo(r *http.Request, args *struct{}, reply *ApiDelegatorNodeInfoReply) error {
	reply.MaximumAllowed = config_nodes.DELEGATES_MAXIMUM
	reply.DelegatesCount = api.wallet.GetDelegatesCount()
	reply.DelegatesFee = config_nodes.DELEGATOR_FEE
	reply.Challenge = api.challenge
	return nil
}

func (api *DelegatorNode) GetDelegatorNodeInfo_http(values url.Values) (interface{}, error) {
	reply := &ApiDelegatorNodeInfoReply{}
	return reply, api.DelegatorNodeInfo(nil, nil, reply)
}

func (api *DelegatorNode) GetDelegatorNodeInfo_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &ApiDelegatorNodeInfoReply{}
	return reply, api.DelegatorNodeInfo(nil, nil, reply)
}
