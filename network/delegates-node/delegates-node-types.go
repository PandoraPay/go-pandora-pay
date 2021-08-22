package delegates_node

import (
	"pandora-pay/network/api/api-common/api_types"
)

type DelegatesNodeInfoRequest struct {
	api_types.APIAccountBaseRequest
}

type DelegatesNodeInfoAnswer struct {
	MaximumAllowed int
	DelegatesCount int
	Exists         bool
}
