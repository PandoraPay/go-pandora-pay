package api_delegator_node

import (
	"net/http"
	"pandora-pay/config/config_nodes"
	"pandora-pay/helpers"
	"sync/atomic"
)

type ApiDelegatorNodeInfoReply struct {
	MaximumAllowed   int            `json:"maximumAllowed" msgpack:"maximumAllowed"`
	DelegatesCount   int            `json:"delegatesCount" msgpack:"delegatesCount"`
	DelegatorFee     uint64         `json:"delegatorFee" msgpack:"delegatorFee"`
	Challenge        helpers.Base64 `json:"challenge" msgpack:"challenge"`
	Blocks           uint64         `json:"blocks" msgpack:"blocks"`
	AcceptCustomKeys bool           `json:"acceptCustomKeys" msgpack:"acceptCustomKeys"`
}

func (api *DelegatorNode) GetDelegatorNodeInfo(r *http.Request, args *struct{}, reply *ApiDelegatorNodeInfoReply) error {
	reply.MaximumAllowed = config_nodes.DELEGATES_MAXIMUM
	reply.AcceptCustomKeys = config_nodes.DELEGATOR_ACCEPT_CUSTOM_KEYS
	reply.DelegatesCount = api.wallet.GetDelegatesCount()
	reply.DelegatorFee = config_nodes.DELEGATOR_FEE
	reply.Challenge = api.challenge
	reply.Blocks = atomic.LoadUint64(&api.chainHeight)
	return nil
}
