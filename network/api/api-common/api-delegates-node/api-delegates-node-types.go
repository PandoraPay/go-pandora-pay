package api_delegates_node

type ApiDelegatesNodeInfoRequest struct {
}

type ApiDelegatesNodeInfoAnswer struct {
	MaximumAllowed int
	DelegatesCount int
}
