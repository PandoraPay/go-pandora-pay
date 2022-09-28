//go:build wasm
// +build wasm

package node_tcp

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/server/node_http"
	"pandora-pay/settings"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
)

type TcpServer struct {
	HttpServer *node_http.HttpServer
}

func NewTcpServer(connectedNodes *connected_nodes.ConnectedNodes, bannedNodes *banned_nodes.BannedNodes, knownNodes *known_nodes.KnownNodes, settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet, txsValidator *txs_validator.TxsValidator, txsBuilder *txs_builder.TxsBuilder) (*TcpServer, error) {

	server := &TcpServer{}
	var err error
	if server.HttpServer, err = node_http.NewHttpServer(chain, settings, connectedNodes, bannedNodes, knownNodes, mempool, wallet, txsValidator, txsBuilder); err != nil {
		return nil, err
	}

	return server, nil
}
