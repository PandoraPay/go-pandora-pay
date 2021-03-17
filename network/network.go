package network

import (
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	node_tcp "pandora-pay/network/server/node-tcp"
	"pandora-pay/network/websockets"
	"pandora-pay/settings"
	"time"
)

type Network struct {
	sockets    *websockets.Websockets
	tcpServer  *node_tcp.TcpServer
	KnownNodes *KnownNodes
}

func (network *Network) execute() {
	go func() {

		gui.Log("connecting to: ")
		time.Sleep(100 * time.Millisecond)

	}()
}

func CreateNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Network {

	sockets := websockets.CreateWebsockets(settings, chain, mempool)
	tcpServer := node_tcp.CreateTcpServer(sockets, settings, chain, mempool)
	knownNodes := CreateKnownNodes()

	for _, seed := range config.NETWORK_SEEDS {
		knownNodes.AddKnownNode(&seed, true)
	}

	network := &Network{
		sockets:    sockets,
		tcpServer:  tcpServer,
		KnownNodes: knownNodes,
	}

	network.execute()

	return network
}
