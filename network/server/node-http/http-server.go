package node_http

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	api_common "pandora-pay/network/api/api-common"
	api_http "pandora-pay/network/api/api-http"
	"pandora-pay/network/api/api-websockets"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
)

type HttpServer struct {
	tcpListener     net.Listener
	Websockets      *websocks.Websockets
	websocketServer *websocks.WebsocketServer
	Api             *api_http.API
	ApiWebsockets   *api_websockets.APIWebsockets
	ApiStore        *api_common.APIStore
	getMap          map[string]func(values *url.Values) (interface{}, error)
}

func (server *HttpServer) get(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	var err error
	var output interface{}

	callback := server.getMap[req.URL.Path]
	if callback != nil {
		arguments := req.URL.Query()
		output, err = callback(&arguments)
	} else {
		err = errors.New("Unknown GET request")
	}
	if err != nil {
		http.Error(w, "Error: "+err.Error(), http.StatusBadRequest)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(output)
	}

}

func (server *HttpServer) initialize() {

	for key, callback := range server.Api.GetMap {
		http.HandleFunc("/"+key, server.get)
		server.getMap["/"+key] = callback
	}

	go func() {
		if err := http.Serve(server.tcpListener, nil); err != nil {
			gui.Error("Error opening HTTP server", err)
		}
		gui.Info("HTTP server")
	}()

}

func CreateHttpServer(tcpListener net.Listener, chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool) (server *HttpServer, err error) {

	apiStore := api_common.CreateAPIStore(chain)
	apiCommon := api_common.CreateAPICommon(mempool, chain, apiStore)

	apiWebsockets := api_websockets.CreateWebsocketsAPI(apiStore, apiCommon, chain, mempool)
	api := api_http.CreateAPI(apiStore, apiCommon, chain)

	websockets := websocks.CreateWebsockets(api, apiWebsockets)

	server = &HttpServer{
		tcpListener:     tcpListener,
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
