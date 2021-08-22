package delegates_node

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/config"
	"pandora-pay/config/config_stake"
	"pandora-pay/helpers"
	node_http "pandora-pay/network/server/node-http"
	"pandora-pay/store"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"pandora-pay/wallet"
)

type DelegatesNode struct {
	wallet     *wallet.Wallet
	httpServer *node_http.HttpServer
}

func (api *DelegatesNode) getDelegatesInfo(request *DelegatesNodeInfoRequest) ([]byte, error) {

	publicKeyHash, err := request.GetPublicKeyHash()
	if err != nil {
		return nil, err
	}

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

	answer := &DelegatesNodeInfoAnswer{
		config.DELEGATES_MAXIMUM,
		api.wallet.GetDelegatesCount(),
		addr == nil,
	}

	return json.Marshal(answer)
}

func CreateDelegatesNode(wallet *wallet.Wallet, httpServer *node_http.HttpServer) (delegates *DelegatesNode) {

	delegates = &DelegatesNode{
		wallet,
		httpServer,
	}

	httpServer.Websockets.ApiWebsockets.GetMap["delegates/info"] = delegates.getDelegatesInfoWebsocket
	httpServer.Api.GetMap["delegates/info"] = delegates.getDelegatesInfoHttp

	return
}
