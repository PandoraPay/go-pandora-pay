package api_delegates_node

import (
	"encoding/binary"
	"encoding/json"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"pandora-pay/wallet"
	"sync"
	"sync/atomic"
	"time"
)

type apiPendingDelegateStakeChange struct {
	delegatePrivateKey *addresses.PrivateKey
	delegatePublicKey  []byte
	publicKey          []byte
	blockHeight        uint64
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
		config.DELEGATES_MAXIMUM,
		api.wallet.GetDelegatesCount(),
		config.DELEGATES_FEES,
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

		accs, err := accsCollection.GetMap(config.NATIVE_TOKEN_FULL)
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

	delegatePrivateKey := addresses.GenerateNewPrivateKey()
	delegatePublicKey := delegatePrivateKey.GeneratePublicKey()

	data, loaded := api.pendingDelegatesStakesChanges.LoadOrStore(string(publicKey), &apiPendingDelegateStakeChange{
		delegatePrivateKey,
		delegatePublicKey,
		publicKey,
		atomic.LoadUint64(&api.chainHeight),
	})
	if loaded {
		pendingDelegateStakeChange := data.(*apiPendingDelegateStakeChange)
		delegatePrivateKey = pendingDelegateStakeChange.delegatePrivateKey
		delegatePublicKey = pendingDelegateStakeChange.delegatePublicKey
	}

	answer := &ApiDelegatesNodeAskAnswer{
		Exists:            false,
		DelegatePublicKey: delegatePublicKey,
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
