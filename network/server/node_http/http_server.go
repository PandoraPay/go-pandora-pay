package node_http

import (
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_http"
	"pandora-pay/network/api/api_websockets"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
)

type HttpServer struct {
	Websockets      *websocks.Websockets
	websocketServer *websocks.WebsocketServer
	Api             *api_http.API
	ApiWebsockets   *api_websockets.APIWebsockets
	ApiStore        *api_common.APIStore
	GetMap          map[string]func(values *url.Values) (interface{}, error)
}

func CreateHttpServer(chain *blockchain.Blockchain, settings *settings.Settings, bannedNodes *banned_nodes.BannedNodes, knownNodes *known_nodes.KnownNodes, mempool *mempool.Mempool, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*HttpServer, error) {

	apiStore := api_common.CreateAPIStore(chain)
	apiCommon, err := api_common.CreateAPICommon(knownNodes, mempool, chain, wallet, transactionsBuilder, apiStore)
	if err != nil {
		return nil, err
	}

	apiWebsockets := api_websockets.CreateWebsocketsAPI(apiStore, apiCommon, chain, settings, mempool)
	api := api_http.CreateAPI(apiStore, apiCommon, chain)

	websockets := websocks.CreateWebsockets(chain, mempool, settings, bannedNodes, api, apiWebsockets)

	server := &HttpServer{
		websocketServer: websocks.CreateWebsocketServer(websockets),
		Websockets:      websockets,
		GetMap:          make(map[string]func(values *url.Values) (interface{}, error)),
		Api:             api,
		ApiWebsockets:   apiWebsockets,
		ApiStore:        apiStore,
	}

	return server, nil
}
