//go:build !wasm
// +build !wasm

package node_http

import (
	"encoding/json"
	"errors"
	"net/http"
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
	case []byte:
		final = v
	default:
		if final, err = json.Marshal(output); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
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

	callback := server.GetMap[req.URL.Path]
	if callback != nil {
		if err = req.ParseForm(); err == nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		output, err = callback(req.Form)
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

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Write(final)
}

func (server *HttpServer) Initialize() {

	for key, callback := range server.Api.GetMap {
		http.HandleFunc("/"+key, server.get)
		server.GetMap["/"+key] = callback
	}

	for key, callback := range server.Api.PostMap {
		http.HandleFunc("/"+key, server.post)
		server.PostMap["/"+key] = callback
	}

}
