package api_delegator_node

import (
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
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

type ApiDelegatorNodeAskAnswer struct {
	Exists                   bool             `json:"exists"`
	DelegateStakingPublicKey helpers.HexBytes `json:"delegateStakingPublicKey"`
}

func (api *DelegatorNode) getDelegatesAsk(request *ApiDelegatorNodeAskRequest) ([]byte, error) {

	publicKey := request.PublicKey

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
		return nil, err
	}

	addr := api.wallet.GetWalletAddressByPublicKey(publicKey)
	if addr != nil {
		return json.Marshal(&ApiDelegatorNodeAskAnswer{
			Exists: true,
		})
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

	answer := &ApiDelegatorNodeAskAnswer{
		Exists:                   false,
		DelegateStakingPublicKey: delegateStakingPublicKey,
	}

	return json.Marshal(answer)
}

func (api *DelegatorNode) GetDelegatorNodeAsk_http(values *url.Values) (interface{}, error) {
	request := &ApiDelegatorNodeAskRequest{}
	var err error
	if challengeSignature := values.Get("challengeSignature"); challengeSignature != "" {
		request.ChallengeSignature, err = hex.DecodeString(challengeSignature)
	} else {
		err = errors.New("'challengeSignature' parameter is missing")
	}
	if err != nil {
		return nil, err
	}

	if publicKey := values.Get("publicKey"); publicKey != "" {
		request.PublicKey, err = hex.DecodeString(publicKey)
	} else {
		err = errors.New("'publicKey' parameter is missing")
	}
	if err != nil {
		return nil, err
	}

	return api.getDelegatesAsk(request)
}

func (api *DelegatorNode) GetDelegatorNodeAsk_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &ApiDelegatorNodeAskRequest{}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.getDelegatesAsk(request)
}
