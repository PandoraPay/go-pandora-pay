package network

import (
	"math/rand"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/network/consensus"
	"pandora-pay/network/known-nodes"
	node_tcp "pandora-pay/network/server/node-tcp"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
	"time"
)

type Network struct {
	tcpServer  *node_tcp.TcpServer
	KnownNodes *known_nodes.KnownNodes
	Websockets *websocks.Websockets
	Consensus  *consensus.Consensus
}

func (network *Network) execute() {

	for {

		var knownNode *known_nodes.KnownNode
		network.KnownNodes.RLock()
		if len(network.KnownNodes.KnownList) > 0 {
			knownNode = network.KnownNodes.KnownList[rand.Intn(len(network.KnownNodes.KnownList))]
		}
		network.KnownNodes.RUnlock()

		_, exists := network.Websockets.AllAddresses.Load(knownNode.UrlHostOnly)
		if !exists {
			gui.Log("connecting to: " + knownNode.UrlStr)

			if knownNode != nil && !exists {
				_, err := websocks.CreateWebsocketClient(network.Websockets, knownNode)
				if err != nil {
					if err.Error() != "Already connected" {
						continue
					}
					gui.Log("error connecting to: " + knownNode.UrlStr)
				} else {
					gui.Log("connected to: " + knownNode.UrlStr)
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
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
		Websockets: tcpServer.HttpServer.Websockets,
		Consensus:  consensus.CreateConsensus(tcpServer.HttpServer, chain, mempool),
	}

	go network.execute()

	return network
}
