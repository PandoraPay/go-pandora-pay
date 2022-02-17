package api_delegator_node

import (
	"errors"
	"net/http"
	"pandora-pay/addresses"
	"pandora-pay/config/config_nodes"
	"pandora-pay/helpers"
	"sync/atomic"
)

type ApiDelegatorNodeAskRequest struct {
	PublicKey                  helpers.Base64 `json:"publicKey" msgpack:"publicKey"`
	ChallengeSignature         helpers.Base64 `json:"challengeSignature" msgpack:"challengeSignature"`
	DelegatedStakingPrivateKey helpers.Base64 `json:"delegatedStakingPrivateKey" msgpack:"delegatedStakingPrivateKey"`
}

type ApiDelegatorNodeAskReply struct {
	Exists                    bool   `json:"exists" msgpack:"exists"`
	DelegatedStakingPublicKey []byte `json:"delegatedStakingPublicKey" msgpack:"delegatedStakingPublicKey"`
}

func (api *DelegatorNode) GetDelegatesAsk(r *http.Request, args *ApiDelegatorNodeAskRequest, reply *ApiDelegatorNodeAskReply, authenticated bool) error {

	if config_nodes.DELEGATOR_REQUIRE_AUTH && !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKey := args.PublicKey

	address, err := addresses.CreateAddr(publicKey, nil, nil, 0, nil)
	if err != nil {
		return nil
	}
	if !address.VerifySignedMessage(api.challenge, args.ChallengeSignature) {
		return errors.New("Challenge was not verified!")
	}

	addr := api.wallet.GetWalletAddressByPublicKey(publicKey, true)
	if addr != nil && addr.PrivateKey == nil {
		reply.Exists = true
		return nil
	}

	var delegatedStakingPrivateKey *addresses.PrivateKey
	if !config_nodes.DELEGATOR_ACCEPT_CUSTOM_KEYS || len(args.DelegatedStakingPrivateKey) == 0 {
		delegatedStakingPrivateKey = addresses.GenerateNewPrivateKey()
	} else {
		if delegatedStakingPrivateKey, err = addresses.CreatePrivateKeyFromSeed(args.DelegatedStakingPrivateKey); err != nil {
			return err
		}
	}

	delegatedStakingPublicKey := delegatedStakingPrivateKey.GeneratePublicKey()

	api.pendingDelegatesStakesChanges.Store(string(publicKey), &PendingDelegateStakeChange{
		delegatedStakingPrivateKey,
		delegatedStakingPublicKey,
		publicKey,
		atomic.LoadUint64(&api.chainHeight),
	})

	reply.DelegatedStakingPublicKey = delegatedStakingPublicKey
	return nil
}
