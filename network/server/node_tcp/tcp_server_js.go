//go:build wasm
// +build wasm

package node_tcp

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/network/server/node_http"
	"pandora-pay/settings"
	"pandora-pay/wallet"
)

type tcpServerType struct {
}

var TcpServer *tcpServerType

func NewTcpServer(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet) error {
	TcpServer = &tcpServerType{}
	return node_http.NewHttpServer(chain, settings, mempool, wallet)
}
