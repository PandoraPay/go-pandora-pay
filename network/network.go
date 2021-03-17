package network

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	node_tcp "pandora-pay/network/server/node-tcp"
	"pandora-pay/network/websockets"
	"pandora-pay/settings"
)

type Network struct {
	sockets   *websockets.Websockets
	tcpServer *node_tcp.TcpServer
}

func CreateNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Network {

	sockets := &websockets.Websockets{}

	tcpServer := node_tcp.CreateTcpServer(sockets, settings, chain, mempool)

	network := &Network{
		sockets:   sockets,
		tcpServer: tcpServer,
	}

	return network
}
