// +build !wasm

package node_http

import (
	"encoding/json"
	"errors"
	"net/http"
)

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

}
