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

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, err.(error).Error(), http.StatusInternalServerError)
		}
	}()

	var err error
	var output interface{}

	callback := server.GetMap[req.URL.Path]
	if callback != nil {
		arguments := req.URL.Query()
		output, err = callback(arguments)
	} else {
		err = errors.New("Unknown request")
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var final []byte
	switch v := output.(type) {
	case string:
		final = []byte(v)
	case helpers.HexBytes:
		final = []byte(hex.EncodeToString(v))
	case []byte:
		final = v
	default:
		if final, err = json.Marshal(output); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(final)
}

func (server *HttpServer) Initialize() {

	for key, callback := range server.Api.GetMap {
		http.HandleFunc("/"+key, server.get)
		server.GetMap["/"+key] = callback
	}

}
