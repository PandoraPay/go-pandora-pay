package network

import (
	"github.com/gorilla/websocket"
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	node_tcp "pandora-pay/network/server/node-tcp"
	"pandora-pay/settings"
)

type Network struct {
	clients           []*websocket.Conn
	serverClients     []*websocket.Conn
	all               []*websocket.Conn
	chanNewConnection chan *websocket.Conn
	tcpServer         *node_tcp.TcpServer
}

func (network *Network) execute() {
	go func() {
		for {
			conn := <-network.chanNewConnection
			network.all = append(network.all, conn)
			network.all = append(network.all, conn)
		}
	}()
}

func CreateNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Network {

	tcpServer := node_tcp.CreateTcpServer(settings, chain, mempool)

	network := &Network{
		clients:           []*websocket.Conn{},
		serverClients:     []*websocket.Conn{},
		all:               []*websocket.Conn{},
		chanNewConnection: make(chan *websocket.Conn),
		tcpServer:         tcpServer,
	}

	network.execute()

	return network
}
