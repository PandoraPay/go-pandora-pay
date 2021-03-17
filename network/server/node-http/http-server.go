package node_http

import (
	"encoding/json"
	"net"
	"net/http"
	"pandora-pay/blockchain"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/network/api"
	node_websocket "pandora-pay/network/server/node-websocket"
)

type HttpServer struct {
	tcpListener net.Listener
	chain       *blockchain.Blockchain
	api         *api.API
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

	callback := server.api.GetMap[req.URL.Path]
	if callback != nil {
		output = callback(req.URL.Query())
	} else {
		panic("Unknown GET request")
	}

}

func (server *HttpServer) initialize() {

	node_websocket.CreateWebsocketServer()

	for key, _ := range server.api.GetMap {
		http.HandleFunc(key, server.get)
	}

	if err := http.Serve(server.tcpListener, nil); err != nil {
		panic(err)
	}

	gui.Info("HTTP server")

}

func CreateHttpServer(tcpListener net.Listener, chain *blockchain.Blockchain, api *api.API) *HttpServer {

	server := &HttpServer{
		tcpListener: tcpListener,
		chain:       chain,
		api:         api,
	}
	server.initialize()

	return server
}
