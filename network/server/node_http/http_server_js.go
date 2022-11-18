//go:build wasm
// +build wasm

package node_http

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/network/api_implementation/api_common"
	"pandora-pay/network/api_implementation/api_websockets"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
	"pandora-pay/wallet"
)

type HttpServer struct {
	Websockets *websocks.Websockets
}

func NewHttpServer(chain *blockchain.Blockchain, settings *settings.Settings, connectedNodes *connected_nodes.ConnectedNodes, bannedNodes *banned_nodes.BannedNodes, knownNodes *known_nodes.KnownNodes, mempool *mempool.Mempool, wallet *wallet.Wallet) (*HttpServer, error) {

	apiStore := api_common.NewAPIStore(chain)
	apiCommon, err := api_common.NewAPICommon(knownNodes, mempool, chain, wallet, apiStore)
	if err != nil {
		return nil, err
	}

	apiWebsockets := api_websockets.NewWebsocketsAPI(apiStore, apiCommon, chain, settings, mempool)
	websockets := websocks.NewWebsockets(chain, mempool, settings, connectedNodes, knownNodes, bannedNodes, apiWebsockets)

	return &HttpServer{
		websockets,
	}, nil
}
