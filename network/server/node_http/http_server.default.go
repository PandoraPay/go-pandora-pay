//go:build !wasm
// +build !wasm

package node_http

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"pandora-pay/helpers"
)

func (server *HttpServer) get(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")

	if req.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	var err error
	var output interface{}

	callback := server.GetMap[req.URL.Path]
	if callback != nil {
		arguments := req.URL.Query()
		output, err = callback(&arguments)
	} else {
		err = errors.New("Unknown GET request")
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		w.Header().Set("Content-Type", "application/json")
		switch v := output.(type) {
		case string:
			w.Write([]byte(v))
		case helpers.HexBytes:
			w.Write([]byte(hex.EncodeToString(v)))
		case []byte:
			w.Write(v)
		default:
			json.NewEncoder(w).Encode(output)
		}
	}

}

func (server *HttpServer) Initialize() {

	for key, callback := range server.Api.GetMap {
		http.HandleFunc("/"+key, server.get)
		server.GetMap["/"+key] = callback
	}

}
