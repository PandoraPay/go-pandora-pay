package node_http

import (
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	api_common "pandora-pay/network/api/api-common"
	api_http "pandora-pay/network/api/api-http"
	"pandora-pay/network/api/api-websockets"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
)

type HttpServer struct {
	Websockets      *websocks.Websockets
	websocketServer *websocks.WebsocketServer
	Api             *api_http.API
	ApiWebsockets   *api_websockets.APIWebsockets
	ApiStore        *api_common.APIStore
	getMap          map[string]func(values *url.Values) (interface{}, error)
}

func CreateHttpServer(chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool) (server *HttpServer, err error) {

	apiStore := api_common.CreateAPIStore(chain)
	apiCommon := api_common.CreateAPICommon(mempool, chain, apiStore)

	apiWebsockets := api_websockets.CreateWebsocketsAPI(apiStore, apiCommon, chain, mempool)
	api := api_http.CreateAPI(apiStore, apiCommon, chain)

	websockets := websocks.CreateWebsockets(api, apiWebsockets)

	server = &HttpServer{
		websocketServer: websocks.CreateWebsocketServer(websockets),
		Websockets:      websockets,
		getMap:          make(map[string]func(values *url.Values) (interface{}, error)),
		Api:             api,
		ApiWebsockets:   apiWebsockets,
		ApiStore:        apiStore,
	}
	server.initialize()

	return
}
