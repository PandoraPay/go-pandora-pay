//go:build wasm
// +build wasm

package node_tcp

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/server/node_http"
	"pandora-pay/settings"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
)

type TcpServer struct {
	Port       string
	HttpServer *node_http.HttpServer
}

func CreateTcpServer(bannedNodes *banned_nodes.BannedNodes, knownNodes *known_nodes.KnownNodes, settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*TcpServer, error) {

	server := &TcpServer{}
	var err error
	if server.HttpServer, err = node_http.CreateHttpServer(chain, settings, bannedNodes, knownNodes, mempool, wallet, transactionsBuilder); err != nil {
		return nil, err
	}

	return server, nil
}
