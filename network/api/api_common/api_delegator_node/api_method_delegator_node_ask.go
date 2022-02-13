package api_delegator_node

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/addresses"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"sync/atomic"
)

type ApiDelegatorNodeAskRequest struct {
	PublicKey          helpers.HexBytes `json:"publicKey" msgpack:"publicKey"`
	ChallengeSignature helpers.HexBytes `json:"challengeSignature" msgpack:"challengeSignature"`
}

type ApiDelegatorNodeAskReply struct {
	Exists                   bool             `json:"exists" msgpack:"exists"`
	DelegateStakingPublicKey helpers.HexBytes `json:"delegateStakingPublicKey" msgpack:"delegateStakingPublicKey"`
}

func (api *DelegatorNode) DelegatesAsk(r *http.Request, args *ApiDelegatorNodeAskRequest, reply *ApiDelegatorNodeAskReply) error {

	publicKey := args.PublicKey

	address, err := addresses.CreateAddr(publicKey, nil, nil, 0, nil)
	if err != nil {
		return nil
	}
	if !address.VerifySignedMessage(args.ChallengeSignature, api.challenge) {
		return errors.New("Challenge was not verified!")
	}

	addr := api.wallet.GetWalletAddressByPublicKey(publicKey, false)
	if addr != nil {
		reply.Exists = true
		return nil
	}

	delegateStakingPrivateKey := addresses.GenerateNewPrivateKey()
	delegateStakingPublicKey := delegateStakingPrivateKey.GeneratePublicKey()

	pendingDelegateStakeChange, loaded := api.pendingDelegatesStakesChanges.LoadOrStore(string(publicKey), &PendingDelegateStakeChange{
		delegateStakingPrivateKey,
		delegateStakingPublicKey,
		publicKey,
		atomic.LoadUint64(&api.chainHeight),
	})

	if loaded {
		delegateStakingPrivateKey = pendingDelegateStakeChange.delegateStakingPrivateKey
		delegateStakingPublicKey = pendingDelegateStakeChange.delegateStakingPublicKey
	}

	reply.DelegateStakingPublicKey = delegateStakingPublicKey
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
