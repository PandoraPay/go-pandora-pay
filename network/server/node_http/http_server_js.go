//go:build wasm
// +build wasm

package node_http

type HttpServer struct {
}

func NewHttpServer() (*HttpServer, error) {
	return &HttpServer{}, nil
}
