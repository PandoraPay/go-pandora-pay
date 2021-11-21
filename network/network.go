package network

import (
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/mempool"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/consensus"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/mempool_sync"
	"pandora-pay/network/server/node_tcp"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/recovery"
	"pandora-pay/settings"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
	"time"
)

type Network struct {
	tcpServer   *node_tcp.TcpServer
	KnownNodes  *known_nodes.KnownNodes
	BannedNodes *banned_nodes.BannedNodes
	MempoolSync *mempool_sync.MempoolSync
	Websockets  *websocks.Websockets
	Consensus   *consensus.Consensus
}

func (network *Network) execute() {

	for {

		if network.Websockets.GetClients() >= config.WEBSOCKETS_NETWORK_CLIENTS_MAX {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		knownNode := network.KnownNodes.GetRandomKnownNode()
		if knownNode == nil {
			continue
		}

		if network.BannedNodes.IsBanned(knownNode.URL) {
			continue //banned already
		}

		_, exists := network.Websockets.AllAddresses.Load(knownNode.URL)
		if !exists {

			if config.DEBUG {
				gui.GUI.Log("connecting to: " + knownNode.URL)
			}

			if knownNode != nil {
				_, err := websocks.CreateWebsocketClient(network.Websockets, knownNode)
				if err != nil {
					if config.DEBUG && err.Error() != "Already connected" {
						if knownNode.IncrementScore(-5, false) {
							network.KnownNodes.RemoveKnownNode(knownNode)
						}
						gui.GUI.Error("error connecting to: "+knownNode.URL, err)
					}
				} else {
					gui.GUI.Log("connected to: " + knownNode.URL)
				}
			}
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (network *Network) continuouslyDownloadMempool() {

	for {

		conn := network.Websockets.GetRandomSocket()
		if conn != nil {

			conn.Send([]byte("chain-get"), nil, nil)

			if config.CONSENSUS == config.CONSENSUS_TYPE_FULL && conn.Handshake.Consensus == config.CONSENSUS_TYPE_FULL {
				network.MempoolSync.DownloadMempool(conn)
			}

		}

		time.Sleep(2000 * time.Millisecond)
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

			//making it async
			recovery.SafeGo(func() {

				conn.Send([]byte("chain-get"), nil, nil)

				if config.CONSENSUS == config.CONSENSUS_TYPE_FULL && conn.Handshake.Consensus == config.CONSENSUS_TYPE_FULL {
					network.MempoolSync.DownloadMempool(conn)
				}

			})

		}
	})
}

func CreateNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*Network, error) {

	knownNodes := known_nodes.CreateKnownNodes()
	for _, seed := range config.NETWORK_SELECTED_SEEDS {
		knownNodes.AddKnownNode(seed.Url, true)
	}

	bannedNodes := banned_nodes.CreateBannedNodes()

	tcpServer, err := node_tcp.CreateTcpServer(bannedNodes, knownNodes, settings, chain, mempool, wallet, transactionsBuilder)
	if err != nil {
		return nil, err
	}

	mempoolSync := mempool_sync.CreateMempoolSync(tcpServer.HttpServer.Websockets)

	network := &Network{
		tcpServer:   tcpServer,
		KnownNodes:  knownNodes,
		BannedNodes: bannedNodes,
		MempoolSync: mempoolSync,
		Websockets:  tcpServer.HttpServer.Websockets,
		Consensus:   consensus.CreateConsensus(tcpServer.HttpServer, chain, mempool),
	}
	tcpServer.HttpServer.Initialize()

	recovery.SafeGo(network.execute)

	if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
		recovery.SafeGo(network.continuouslyDownloadMempool)
	}

	recovery.SafeGo(network.syncNewConnections)

	return network, nil
}
