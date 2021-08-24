package api_delegates_node

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config"
	"pandora-pay/config/config_stake"
	"pandora-pay/cryptography"
	"pandora-pay/cryptography/ecdsa"
	"pandora-pay/helpers"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"pandora-pay/wallet"
	"sync"
	"sync/atomic"
	"time"
)

type apiPendingDelegateStakeChange struct {
	delegatePrivateKey    *addresses.PrivateKey
	delegatePublicKeyHash []byte
	publicKeyHash         []byte
	publicKey             []byte
	blockHeight           uint64
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

	publicKey, err := ecdsa.EcrecoverCompressed(api.challenge, request.ChallengeSignature)
	if err != nil {
		return nil, err
	}

	publicKeyHash := cryptography.ComputePublicKeyHash(publicKey)

	var chainHeight uint64
	var acc *account.Account
	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		chainHeight, _ = binary.Uvarint(reader.Get("chainHeight"))
		acc, err = accounts.NewAccounts(reader).GetAccount(publicKeyHash, chainHeight)
		return
	}); err != nil {
		return nil, err
	}

	amount, err := acc.ComputeDelegatedStakeAvailable(chainHeight)
	if err != nil {
		return nil, err
	}

	amount2 := acc.GetAvailableBalance(config.NATIVE_TOKEN)
	if err = helpers.SafeUint64Add(&amount, amount2); err != nil {
		return nil, err
	}

	requiredStake := config_stake.GetRequiredStake(chainHeight)
	if amount < requiredStake {
		return nil, errors.New("You will not enought to stake")
	}

	addr := api.wallet.GetWalletAddressByPublicKeyHash(publicKeyHash)
	if addr != nil {
		return json.Marshal(&ApiDelegatesNodeAskAnswer{
			Exists: true,
		})
	}

	delegatePrivateKey := addresses.GenerateNewPrivateKey()
	delegatePublicKeyHash, err := delegatePrivateKey.GeneratePublicKeyHash()
	if err != nil {
		return nil, err
	}

	data, loaded := api.pendingDelegatesStakesChanges.LoadOrStore(string(publicKeyHash), &apiPendingDelegateStakeChange{
		delegatePrivateKey,
		delegatePublicKeyHash,
		publicKeyHash,
		publicKey,
		atomic.LoadUint64(&api.chainHeight),
	})
	if loaded {
		pendingDelegateStakeChange := data.(*apiPendingDelegateStakeChange)
		delegatePrivateKey = pendingDelegateStakeChange.delegatePrivateKey
		delegatePublicKeyHash = pendingDelegateStakeChange.delegatePublicKeyHash
	}

	answer := &ApiDelegatesNodeAskAnswer{
		Exists:                false,
		DelegatePublicKeyHash: delegatePublicKeyHash,
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
