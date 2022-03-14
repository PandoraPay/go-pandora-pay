package api_delegator_node

import (
	"errors"
	"net/http"
	"pandora-pay/addresses"
	"pandora-pay/config/config_nodes"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type ApiDelegatorNodeAskRequest struct {
	PublicKey                  helpers.Base64 `json:"publicKey" msgpack:"publicKey"`
	ChallengeSignature         helpers.Base64 `json:"challengeSignature" msgpack:"challengeSignature"`
	DelegatedStakingPrivateKey helpers.Base64 `json:"delegatedStakingPrivateKey,omitempty" msgpack:"delegatedStakingPrivateKey,omitempty"`
}

type ApiDelegatorNodeAskReply struct {
	Exists                    bool   `json:"exists" msgpack:"exists"`
	DelegatedStakingPublicKey []byte `json:"delegatedStakingPublicKey" msgpack:"delegatedStakingPublicKey"`
}

func (api *DelegatorNode) DelegatorAsk(r *http.Request, args *ApiDelegatorNodeAskRequest, reply *ApiDelegatorNodeAskReply, authenticated bool) error {

	if config_nodes.DELEGATOR_REQUIRE_AUTH && !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKey := args.PublicKey

	address, err := addresses.CreateAddr(publicKey, false, nil, nil, nil, 0, nil)
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
		key := cryptography.SHA3(append(args.DelegatedStakingPrivateKey, api.secret...))
		delegatedStakingPrivateKey = &addresses.PrivateKey{key}
	} else {
		if delegatedStakingPrivateKey, err = addresses.CreatePrivateKeyFromSeed(args.DelegatedStakingPrivateKey); err != nil {
			return err
		}
	}

	delegatedStakingPublicKey := delegatedStakingPrivateKey.GeneratePublicKey()

	reply.DelegatedStakingPublicKey = delegatedStakingPublicKey
	return nil
}
