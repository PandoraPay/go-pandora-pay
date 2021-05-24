// +build wasm

package node_tcp

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	node_http "pandora-pay/network/server/node-http"
	"pandora-pay/settings"
)

type TcpServer struct {
	Port       string
	HttpServer *node_http.HttpServer
}

func CreateTcpServer(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) (server *TcpServer, err error) {
	server = &TcpServer{}

	server.HttpServer, err = node_http.CreateHttpServer(chain, settings, mempool)

	return
}
