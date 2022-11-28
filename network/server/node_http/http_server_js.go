//go:build wasm
// +build wasm

package node_http

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/network/api_implementation/api_common"
	"pandora-pay/network/api_implementation/api_websockets"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
	"pandora-pay/wallet"
)

type httpServerType struct {
	ApiWebsockets *api_websockets.APIWebsockets
}

var HttpServer *httpServerType

func NewHttpServer(chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool, wallet *wallet.Wallet) error {

	apiStore := api_common.NewAPIStore(chain)
	apiCommon, err := api_common.NewAPICommon(mempool, chain, wallet, apiStore)
	if err != nil {
		return err
	}

	apiWebsockets := api_websockets.NewWebsocketsAPI(apiStore, apiCommon, chain, settings, mempool)
	websocks.NewWebsockets(chain, mempool, settings, apiWebsockets.GetMap)

	HttpServer = &httpServerType{
		apiWebsockets,
	}

	return nil
}
