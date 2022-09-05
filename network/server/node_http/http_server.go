package node_http

import (
	"io"
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_http"
	"pandora-pay/network/api/api_websockets"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/server/node_http_rpc"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
)

type HttpServer struct {
	Websockets      *websocks.Websockets
	websocketServer *websocks.WebsocketServer
	Api             *api_http.API
	ApiWebsockets   *api_websockets.APIWebsockets
	ApiStore        *api_common.APIStore
	GetMap          map[string]func(values url.Values) (any, error)
	PostMap         map[string]func(values io.ReadCloser) (any, error)
}

func NewHttpServer(chain *blockchain.Blockchain, settings *settings.Settings, connectedNodes *connected_nodes.ConnectedNodes, bannedNodes *banned_nodes.BannedNodes, knownNodes *known_nodes.KnownNodes, mempool *mempool.Mempool, wallet *wallet.Wallet, txsValidator *txs_validator.TxsValidator, txsBuilder *txs_builder.TxsBuilder) (*HttpServer, error) {

	apiStore := api_common.NewAPIStore(chain)
	apiCommon, err := api_common.NewAPICommon(knownNodes, mempool, chain, wallet, txsValidator, txsBuilder, apiStore)
	if err != nil {
		return nil, err
	}

	apiWebsockets := api_websockets.NewWebsocketsAPI(apiStore, apiCommon, chain, settings, mempool, txsValidator)
	api := api_http.NewAPI(apiStore, apiCommon, chain)

	websockets := websocks.NewWebsockets(chain, mempool, settings, connectedNodes, knownNodes, bannedNodes, api, apiWebsockets)

	server := &HttpServer{
		websocketServer: websocks.NewWebsocketServer(websockets, connectedNodes, knownNodes),
		Websockets:      websockets,
		GetMap:          make(map[string]func(values url.Values) (any, error)),
		PostMap:         make(map[string]func(values io.ReadCloser) (any, error)),
		Api:             api,
		ApiWebsockets:   apiWebsockets,
		ApiStore:        apiStore,
	}

	if err = node_http_rpc.InitializeRPC(apiCommon); err != nil {
		return nil, err
	}

	return server, nil
}
