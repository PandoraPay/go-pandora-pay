package node_http

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api"
	"pandora-pay/network/api-websockets"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
)

type HttpServer struct {
	tcpListener     net.Listener
	Websockets      *websocks.Websockets
	websocketServer *websocks.WebsocketServer
	Api             *api.API
	ApiWebsockets   *api_websockets.APIWebsockets
	getMap          map[string]func(values *url.Values) interface{}
}

func (server *HttpServer) get(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	var output interface{}

	defer func() {
		if err := helpers.ConvertRecoverError(recover()); err != nil {
			http.Error(w, "Error"+err.Error(), http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(output)
		}
	}()

	callback := server.getMap[req.URL.Path]
	if callback != nil {
		arguments := req.URL.Query()
		output = callback(&arguments)
	} else {
		panic("Unknown GET request")
	}

}

func (server *HttpServer) initialize() {

	for key, callback := range server.Api.GetMap {
		http.HandleFunc("/"+key, server.get)
		server.getMap["/"+key] = callback
	}

	go func() {
		if err := http.Serve(server.tcpListener, nil); err != nil {
			panic(err)
		}
		gui.Info("HTTP server")
	}()

}

func CreateHttpServer(tcpListener net.Listener, chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool) *HttpServer {

	apiWebsockets := api_websockets.CreateWebsocketsAPI(chain, settings, mempool)
	api := api.CreateAPI(chain, settings, mempool)

	websockets := websocks.CreateWebsockets(api, apiWebsockets)

	server := &HttpServer{
		tcpListener:     tcpListener,
		websocketServer: websocks.CreateWebsocketServer(websockets),
		Websockets:      websockets,
		getMap:          make(map[string]func(values *url.Values) interface{}),
		Api:             api,
		ApiWebsockets:   apiWebsockets,
	}
	server.initialize()

	return server
}
