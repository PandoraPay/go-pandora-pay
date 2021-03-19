package network

import (
	"math/rand"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/network/known-nodes"
	node_tcp "pandora-pay/network/server/node-tcp"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
	"time"
)

type Network struct {
	tcpServer  *node_tcp.TcpServer
	KnownNodes *known_nodes.KnownNodes
	websockets *websocks.Websockets
}

func (network *Network) execute() {
	go func() {

		gui.Log("connecting to: ")

		var knownNode *known_nodes.KnownNode
		network.KnownNodes.RLock()
		if len(network.KnownNodes.KnownList) > 0 {
			knownNode = network.KnownNodes.KnownList[rand.Intn(len(network.KnownNodes.KnownList))]
		}
		network.KnownNodes.RUnlock()

		_, exists := network.websockets.AllAddresses.Load(knownNode.Url.String())

		if knownNode != nil && !exists {
			_, err := websocks.CreateWebsocketClient(network.websockets, knownNode)
			if err != nil {
			}
		}

		time.Sleep(100 * time.Millisecond)

	}()
}

func CreateNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) *Network {

	tcpServer := node_tcp.CreateTcpServer(settings, chain, mempool)
	knownNodes := known_nodes.CreateKnownNodes()

	for _, seed := range config.NETWORK_SEEDS {
		knownNodes.AddKnownNode(&seed, true)
	}

	network := &Network{
		tcpServer:  tcpServer,
		KnownNodes: knownNodes,
		websockets: tcpServer.HttpServer.Websockets,
	}

	network.execute()

	return network
}
