package network

import (
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
	"pandora-pay/recovery"
	"pandora-pay/settings"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/wallet"
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

		if network.Websockets.GetClients() >= config.WEBSOCKETS_NETWORK_CLIENTS_MAX {
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		knownNode := network.KnownNodes.GetRandomKnownNode()
		if knownNode == nil {
			continue
		}

		if knownNode.Url.Hostname() == "127.0.0.1" && knownNode.Url.Port() == network.tcpServer.Port {
			continue //skip connecting to myself
		}

		_, exists := network.Websockets.AllAddresses.Load(knownNode.UrlStr)
		if !exists {

			if config.DEBUG {
				gui.GUI.Log("connecting to: " + knownNode.UrlStr)
			}

			if knownNode != nil {
				_, err := websocks.CreateWebsocketClient(network.Websockets, knownNode)
				if err != nil {
					if config.DEBUG && err.Error() != "Already connected" {
						gui.GUI.Error("error connecting to: "+knownNode.UrlStr, err)
					}
				} else {
					gui.GUI.Log("connected to: " + knownNode.UrlStr)
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (network *Network) syncNewConnections() {
	recovery.SafeGo(func() {

		cn := network.Websockets.UpdateNewConnectionMulticast.AddListener()
		defer network.Websockets.UpdateNewConnectionMulticast.RemoveChannel(cn)

		for {

			data, ok := <-cn
			if !ok {
				return
			}
			conn := data.(*connection.AdvancedConnection)

			conn.Send([]byte("chain-get"), nil)

			if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
				network.MempoolSync.DownloadMempool(conn)
			}
		}
	})
}

func CreateNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*Network, error) {

	tcpServer, err := node_tcp.CreateTcpServer(settings, chain, mempool, wallet, transactionsBuilder)
	if err != nil {
		return nil, err
	}

	knownNodes := known_nodes.CreateKnownNodes()
	for _, seed := range config.NETWORK_SELECTED_SEEDS {
		knownNodes.AddKnownNode(&seed, true)
	}

	mempoolSync := mempool_sync.CreateMempoolSync(tcpServer.HttpServer.Websockets)

	network := &Network{
		tcpServer:   tcpServer,
		KnownNodes:  knownNodes,
		MempoolSync: mempoolSync,
		Websockets:  tcpServer.HttpServer.Websockets,
		Consensus:   consensus.CreateConsensus(tcpServer.HttpServer, chain, mempool),
	}

	recovery.SafeGo(network.execute)
	recovery.SafeGo(network.syncNewConnections)

	return network, nil
}
