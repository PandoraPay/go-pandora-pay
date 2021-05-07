package network

import (
	"math/rand"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/network/consensus"
	"pandora-pay/network/known-nodes"
	mempool_sync "pandora-pay/network/mempool-sync"
	node_tcp "pandora-pay/network/server/node-tcp"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/settings"
	"time"
)

type Network struct {
	tcpServer   *node_tcp.TcpServer
	KnownNodes  *known_nodes.KnownNodes
	MempoolSync *mempool_sync.MempoolSync
	Websockets  *websocks.Websockets
	Consensus   *consensus.Consensus
}

func (network *Network) execute() {

	for {

		var knownNode *known_nodes.KnownNode
		knownList := network.KnownNodes.KnownList.Load().([]*known_nodes.KnownNode)
		if len(knownList) > 0 {
			knownNode = knownList[rand.Intn(len(knownList))]
		}

		if knownNode.Url.Hostname() == "127.0.0.1" && knownNode.Url.Port() == network.tcpServer.Port {
			continue //skip connecting to myself
		}

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

func (network *Network) syncNewConnections() {
	go func() {
		for {
			data, ok := <-network.Websockets.UpdateNewConnectionMulticast.AddListener()
			if !ok {
				return
			}
			conn := data.(*connection.AdvancedConnection)

			conn.Send([]byte("chain-get"), nil)

			network.MempoolSync.DownloadMempool(conn)
		}
	}()
}

func CreateNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool) (network *Network, err error) {

	tcpServer, err := node_tcp.CreateTcpServer(settings, chain, mempool)
	if err != nil {
		return
	}

	knownNodes := known_nodes.CreateKnownNodes()
	for _, seed := range config.NETWORK_SEEDS {
		knownNodes.AddKnownNode(&seed, true)
	}

	mempoolSync := mempool_sync.CreateMempoolSync(tcpServer.HttpServer.Websockets)

	network = &Network{
		tcpServer:   tcpServer,
		KnownNodes:  knownNodes,
		MempoolSync: mempoolSync,
		Websockets:  tcpServer.HttpServer.Websockets,
		Consensus:   consensus.CreateConsensus(tcpServer.HttpServer, chain, mempool),
	}

	go network.execute()
	go network.syncNewConnections()

	return
}
