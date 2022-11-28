//go:build !wasm
// +build !wasm

package node_http

import (
	"encoding/json"
	"errors"
	"github.com/rs/cors"
	"io"
	"net/http"
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/network/api_implementation/api_common"
	"pandora-pay/network/api_implementation/api_http"
	"pandora-pay/network/api_implementation/api_websockets"
	"pandora-pay/network/network_config"
	"pandora-pay/network/server/node_http_rpc"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
	"pandora-pay/wallet"
)

type httpServerType struct {
	Api           *api_http.API
	ApiWebsockets *api_websockets.APIWebsockets
	ApiStore      *api_common.APIStore
	GetMap        map[string]func(values url.Values) (any, error)
	PostMap       map[string]func(values io.ReadCloser) (any, error)
}

var HttpServer *httpServerType

func (this *httpServerType) get(w http.ResponseWriter, req *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, err.(error).Error(), http.StatusInternalServerError)
		}
	}()

	var err error
	var output interface{}

	callback := this.GetMap[req.URL.Path]
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

func (this *httpServerType) post(w http.ResponseWriter, req *http.Request) {

	defer func() {
		if err := recover(); err != nil {
			http.Error(w, err.(error).Error(), http.StatusInternalServerError)
		}
	}()

	var err error
	var output interface{}

	callback := this.PostMap[req.URL.Path]
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

func (this *httpServerType) GetHttpHandler() *http.Handler {

	mux := http.NewServeMux()

	mux.HandleFunc("/ws", websocks.Websockets.HandleUpgradeConnection)

	for key, filepath := range network_config.STATIC_FILES {
		fs := http.FileServer(http.Dir(filepath))
		mux.Handle(key, http.StripPrefix(key, fs))
	}

	for key, callback := range this.Api.GetMap {
		mux.HandleFunc("/"+key, this.get)
		this.GetMap["/"+key] = callback
	}

	for key, callback := range this.Api.PostMap {
		mux.HandleFunc("/"+key, this.post)
		this.PostMap["/"+key] = callback
	}

	handler := cors.AllowAll().Handler(mux)
	return &handler
}

func NewHttpServer(chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool, wallet *wallet.Wallet) error {

	apiStore := api_common.NewAPIStore(chain)
	apiCommon, err := api_common.NewAPICommon(mempool, chain, wallet, apiStore)
	if err != nil {
		return err
	}

	apiWebsockets := api_websockets.NewWebsocketsAPI(apiStore, apiCommon, chain, settings, mempool)
	api := api_http.NewAPI(apiStore, apiCommon, chain)

	websocks.NewWebsockets(chain, mempool, settings, apiWebsockets.GetMap)

	HttpServer = &httpServerType{
		api,
		apiWebsockets,
		apiStore,
		make(map[string]func(values url.Values) (any, error)),
		make(map[string]func(values io.ReadCloser) (any, error)),
	}

	if err = node_http_rpc.InitializeRPC(apiCommon); err != nil {
		return err
	}

	return nil
}
