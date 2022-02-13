package api_delegator_node

import (
	"errors"
	"fmt"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/addresses"
	"pandora-pay/config/globals"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"sync/atomic"
)

type ApiDelegatorNodeAskRequest struct {
	PublicKey                  helpers.HexBytes `json:"publicKey" msgpack:"publicKey"`
	ChallengeSignature         helpers.HexBytes `json:"challengeSignature" msgpack:"challengeSignature"`
	DelegatedStakingPrivateKey helpers.HexBytes `json:"delegatedStakingPrivateKey" msgpack:"delegatedStakingPrivateKey"`
}

type ApiDelegatorNodeAskReply struct {
	Exists                    bool             `json:"exists" msgpack:"exists"`
	DelegatedStakingPublicKey helpers.HexBytes `json:"delegatedStakingPublicKey" msgpack:"delegatedStakingPublicKey"`
}

func (api *DelegatorNode) DelegatesAsk(r *http.Request, args *ApiDelegatorNodeAskRequest, reply *ApiDelegatorNodeAskReply) error {

	publicKey := args.PublicKey

	address, err := addresses.CreateAddr(publicKey, nil, nil, 0, nil)
	if err != nil {
		return nil
	}
	if !address.VerifySignedMessage(api.challenge, args.ChallengeSignature) {
		return errors.New("Challenge was not verified!")
	}

	addr := api.wallet.GetWalletAddressByPublicKey(publicKey, false)
	if addr != nil {
		reply.Exists = true
		return nil
	}

	var delegatedStakingPrivateKey *addresses.PrivateKey
	if globals.Arguments["--delegator-accept-custom-keys"] != "true" {
		delegatedStakingPrivateKey = addresses.GenerateNewPrivateKey()
	} else {
		if len(args.DelegatedStakingPrivateKey) != cryptography.PrivateKeySize {
			return fmt.Errorf("delegatedStakingPrivateKey must be of size %d", cryptography.PrivateKeySize)
		}
		delegatedStakingPrivateKey = &addresses.PrivateKey{args.DelegatedStakingPrivateKey}
	}

	delegatedStakingPublicKey := delegatedStakingPrivateKey.GeneratePublicKey()

	pendingDelegateStakeChange, loaded := api.pendingDelegatesStakesChanges.LoadOrStore(string(publicKey), &PendingDelegateStakeChange{
		delegatedStakingPrivateKey,
		delegatedStakingPublicKey,
		publicKey,
		atomic.LoadUint64(&api.chainHeight),
	})

	if loaded {
		delegatedStakingPrivateKey = pendingDelegateStakeChange.delegateStakingPrivateKey
		delegatedStakingPublicKey = pendingDelegateStakeChange.delegateStakingPublicKey
	}

	reply.DelegatedStakingPublicKey = delegatedStakingPublicKey
	return nil
}

func (api *DelegatorNode) GetDelegatorNodeAsk_http(values url.Values) (interface{}, error) {
	args := &ApiDelegatorNodeAskRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &ApiDelegatorNodeAskReply{}
	return reply, api.DelegatesAsk(nil, args, reply)
}

func (api *DelegatorNode) GetDelegatorNodeAsk_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &ApiDelegatorNodeAskRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &ApiDelegatorNodeAskReply{}
	return reply, api.DelegatesAsk(nil, args, reply)
}
