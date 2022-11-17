package api_delegator_node

import (
	"net/http"
	"pandora-pay/config/config_nodes"
	"sync/atomic"
)

type ApiDelegatorNodeInfoReply struct {
	MaximumAllowed int    `json:"maximumAllowed" msgpack:"maximumAllowed"`
	DelegatesCount int    `json:"delegatesCount" msgpack:"delegatesCount"`
	Blocks         uint64 `json:"blocks" msgpack:"blocks"`
}

func (api *DelegatorNode) GetDelegatorNodeInfo(r *http.Request, args *struct{}, reply *ApiDelegatorNodeInfoReply) error {
	reply.MaximumAllowed = config_nodes.DELEGATES_MAXIMUM
	reply.DelegatesCount = api.wallet.GetDelegatesCount()
	reply.Blocks = atomic.LoadUint64(&api.chainHeight)
	return nil
}
