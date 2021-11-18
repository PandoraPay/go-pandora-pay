package api_delegates_node

import (
	"pandora-pay/helpers"
)

type ApiDelegatesNodeInfoRequest struct {
}

type ApiDelegatesNodeInfoAnswer struct {
	MaximumAllowed int              `json:"maximumAllowed"`
	DelegatesCount int              `json:"delegatesCount"`
	DelegatesFee   uint64           `json:"delegatesFee"`
	Challenge      helpers.HexBytes `json:"challenge"`
}

type ApiDelegatesNodeAskRequest struct {
	PublicKey          helpers.HexBytes `json:"publicKey"`
	ChallengeSignature helpers.HexBytes `json:"challengeSignature"`
}

type ApiDelegatesNodeAskAnswer struct {
	Exists                   bool             `json:"exists"`
	DelegateStakingPublicKey helpers.HexBytes `json:"delegateStakingPublicKey"`
}
