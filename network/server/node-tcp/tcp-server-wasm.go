//go:build wasm
// +build wasm

package node_tcp

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	node_http "pandora-pay/network/server/node-http"
	"pandora-pay/settings"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
)

type TcpServer struct {
	Port       string
	HttpServer *node_http.HttpServer
}

func CreateTcpServer(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*TcpServer, error) {

	server := &TcpServer{}
	var err error
	if server.HttpServer, err = node_http.CreateHttpServer(chain, settings, nil, mempool, wallet, transactionsBuilder); err != nil {
		return nil, err
	}

	return server, nil
}
