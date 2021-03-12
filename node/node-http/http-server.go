package node_http

import (
	"encoding/json"
	"net/http"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers"
)

type HttpServer struct {
	getMaps map[string]func(w http.ResponseWriter, req *http.Request) interface{}
}

func (server *HttpServer) get(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	var output interface{}

	defer func() {
		if err := recover(); err != nil {
			error := helpers.ConvertRecoverError(err)
			http.Error(w, "Error"+error.Error(), http.StatusBadRequest)
		} else {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(output)
		}
	}()

	callback := server.getMaps[req.URL.Path]
	if callback != nil {
		output = callback(w, req)
	} else {
		panic("Unknown GET request")
	}

}

func (server *HttpServer) info(w http.ResponseWriter, req *http.Request) interface{} {
	return struct {
		Name        string
		Version     string
		Network     uint64
		CPU_THREADS int
	}{
		Name:        config.NAME,
		Version:     config.VERSION,
		Network:     config.NETWORK_SELECTED,
		CPU_THREADS: config.CPU_THREADS,
	}
}

func (server *HttpServer) ping(w http.ResponseWriter, req *http.Request) interface{} {
	return struct {
		Ping string
	}{Ping: "Pong"}
}

func (server *HttpServer) initialize() {

	server.getMaps = make(map[string]func(w http.ResponseWriter, req *http.Request) interface{})
	server.getMaps["/"] = server.info
	server.getMaps["/ping"] = server.ping

	for key, _ := range server.getMaps {
		http.HandleFunc(key, server.get)
	}

	port := "8090"
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}

	gui.Info("HTTP server opened on port ", port)

}

func CreateHttpServer() *HttpServer {

	server := &HttpServer{}
	server.initialize()

	return server
}
