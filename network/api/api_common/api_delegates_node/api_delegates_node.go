package api_delegates_node

import (
	"encoding/binary"
	"encoding/json"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/config/config_coins"
	"pandora-pay/config/config_nodes"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet"
	"sync"
	"sync/atomic"
	"time"
)

type apiPendingDelegateStakeChange struct {
	delegateStakingPrivateKey *addresses.PrivateKey
	delegateStakingPublicKey  []byte
	publicKey                 []byte
	blockHeight               uint64
}

type APIDelegatesNode struct {
	challenge                     []byte
	chainHeight                   uint64    //use atomic
	pendingDelegatesStakesChanges *sync.Map //*apiPendingDelegateStakeChange
	ticker                        *time.Ticker
	wallet                        *wallet.Wallet
	chain                         *blockchain.Blockchain
}

func (api *APIDelegatesNode) getDelegatesInfo(request *ApiDelegatesNodeInfoRequest) ([]byte, error) {

	answer := &ApiDelegatesNodeInfoAnswer{
		config_nodes.DELEGATES_MAXIMUM,
		api.wallet.GetDelegatesCount(),
		config_nodes.DELEGATES_FEE,
		api.challenge,
	}

	return json.Marshal(answer)
}

func (api *APIDelegatesNode) getDelegatesAsk(request *ApiDelegatesNodeAskRequest) ([]byte, error) {

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
		return json.Marshal(&ApiDelegatesNodeAskAnswer{
			Exists: true,
		})
	}

	delegateStakingPrivateKey := addresses.GenerateNewPrivateKey()
	delegateStakingPublicKey := delegateStakingPrivateKey.GeneratePublicKey()

	data, loaded := api.pendingDelegatesStakesChanges.LoadOrStore(string(publicKey), &apiPendingDelegateStakeChange{
		delegateStakingPrivateKey,
		delegateStakingPublicKey,
		publicKey,
		atomic.LoadUint64(&api.chainHeight),
	})
	if loaded {
		pendingDelegateStakeChange := data.(*apiPendingDelegateStakeChange)
		delegateStakingPrivateKey = pendingDelegateStakeChange.delegateStakingPrivateKey
		delegateStakingPublicKey = pendingDelegateStakeChange.delegateStakingPublicKey
	}

	answer := &ApiDelegatesNodeAskAnswer{
		Exists:                   false,
		DelegateStakingPublicKey: delegateStakingPublicKey,
	}

	return json.Marshal(answer)
}

func CreateDelegatesNode(chain *blockchain.Blockchain, wallet *wallet.Wallet) (delegates *APIDelegatesNode) {

	challenge := helpers.RandomBytes(cryptography.HashSize)

	delegates = &APIDelegatesNode{
		challenge,
		0,
		&sync.Map{},
		nil,
		wallet,
		chain,
	}

	delegates.execute()

	return
}
