//go:build !wasm
// +build !wasm

package node_http

import (
	"encoding/json"
	"errors"
	"github.com/rs/cors"
	"net/http"
	"net/url"
	"pandora-pay/config"
)

func (server *HttpServer) get(w http.ResponseWriter, req *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, err.(error).Error(), http.StatusInternalServerError)
		}
	}()

	var err error
	var output interface{}

	callback := server.GetMap[req.URL.Path]
	if callback != nil {

		var args url.Values
		if args, err = url.ParseQuery(req.URL.RawQuery); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		output, err = callback(args)
	} else {
		err = errors.New("Unknown request")
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var final []byte
	switch v := output.(type) {
	case string:
		final = []byte(v)
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

func (server *HttpServer) post(w http.ResponseWriter, req *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, err.(error).Error(), http.StatusInternalServerError)
		}
	}()

	var err error
	var output interface{}

	callback := server.PostMap[req.URL.Path]
	if callback != nil {
		output, err = callback(req.Body)
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

func (server *HttpServer) GetHttpHandler() *http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", server.websocketServer.HandleUpgradeConnection)

	if config.FAUCET_TESTNET_ENABLED {
		fs := http.FileServer(http.Dir("../../../static/challenge"))
		mux.Handle("/static/challenge/", http.StripPrefix("/static/challenge/", fs))
	}

	for key, callback := range server.Api.GetMap {
		mux.HandleFunc("/"+key, server.get)
		server.GetMap["/"+key] = callback
	}

	for key, callback := range server.Api.PostMap {
		mux.HandleFunc("/"+key, server.post)
		server.PostMap["/"+key] = callback
	}

	handler := cors.AllowAll().Handler(mux)
	return &handler
}
