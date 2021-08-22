package api_delegates_node

import (
	"pandora-pay/helpers"
)

type ApiDelegatesNodeInfoRequest struct {
}

type ApiDelegatesNodeInfoAnswer struct {
	MaximumAllowed int              `json:"maximumAllowed"`
	DelegatesCount int              `json:"delegatesCount"`
	Challenge      helpers.HexBytes `json:"challenge"`
}

type ApiDelegatesNodeAskRequest struct {
	ChallengeSignature helpers.HexBytes `json:"challengeSignature"`
}

type ApiDelegatesNodeAskAnswer struct {
	Exists bool `json:"exists"`
}
