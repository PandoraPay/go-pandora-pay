package api_delegator_node

import (
	"encoding/binary"
	"encoding/json"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"sync/atomic"
)

type ApiDelegatorNodeAskRequest struct {
	PublicKey          helpers.HexBytes `json:"publicKey"`
	ChallengeSignature helpers.HexBytes `json:"challengeSignature"`
}

type ApiDelegatorNodeAskReply struct {
	Exists                   bool             `json:"exists"`
	DelegateStakingPublicKey helpers.HexBytes `json:"delegateStakingPublicKey"`
}

func (api *DelegatorNode) DelegatesAsk(r *http.Request, args *ApiDelegatorNodeAskRequest, reply *ApiDelegatorNodeAskReply) error {

	publicKey := args.PublicKey

	var chainHeight uint64
	var acc *account.Account
	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		accsCollection := accounts.NewAccountsCollection(reader)

		accs, err := accsCollection.GetMap(config_coins.NATIVE_ASSET_FULL)
		if err != nil {
			return
		}
		acc, err = accs.GetAccount(publicKey)
		return

	}); err != nil {
		return err
	}

	addr := api.wallet.GetWalletAddressByPublicKey(publicKey)
	if addr != nil {
		reply.Exists = true
		return nil
	}

	delegateStakingPrivateKey := addresses.GenerateNewPrivateKey()
	delegateStakingPublicKey := delegateStakingPrivateKey.GeneratePublicKey()

	data, loaded := api.pendingDelegatesStakesChanges.LoadOrStore(string(publicKey), &pendingDelegateStakeChange{
		delegateStakingPrivateKey,
		delegateStakingPublicKey,
		publicKey,
		atomic.LoadUint64(&api.chainHeight),
	})
	if loaded {
		pendingDelegateStakeChange := data.(*pendingDelegateStakeChange)
		delegateStakingPrivateKey = pendingDelegateStakeChange.delegateStakingPrivateKey
		delegateStakingPublicKey = pendingDelegateStakeChange.delegateStakingPublicKey
	}

	reply.DelegateStakingPublicKey = delegateStakingPublicKey
	return nil
}

func (api *DelegatorNode) GetDelegatorNodeAsk_http(values url.Values) (interface{}, error) {
	args := &ApiDelegatorNodeAskRequest{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &ApiDelegatorNodeAskReply{}
	return reply, api.DelegatesAsk(nil, args, reply)
}

func (api *DelegatorNode) GetDelegatorNodeAsk_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &ApiDelegatorNodeAskRequest{}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &ApiDelegatorNodeAskReply{}
	return reply, api.DelegatesAsk(nil, args, reply)
}
