package network

import (
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/known_nodes_sync"
	"pandora-pay/network/mempool_sync"
	"pandora-pay/network/server/node_tcp"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
)

type Network struct {
	tcpServer      *node_tcp.TcpServer
	Websockets     *websocks.Websockets
	KnownNodes     *known_nodes.KnownNodes
	BannedNodes    *banned_nodes.BannedNodes
	MempoolSync    *mempool_sync.MempoolSync
	KnownNodesSync *known_nodes_sync.KnownNodesSync
}

func NewNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*Network, error) {

	knownNodes := known_nodes.NewKnownNodes()
	for _, seed := range config.NETWORK_SELECTED_SEEDS {
		knownNodes.AddKnownNode(seed.Url, true)
	}

	bannedNodes := banned_nodes.NewBannedNodes()

	tcpServer, err := node_tcp.NewTcpServer(bannedNodes, knownNodes, settings, chain, mempool, wallet, transactionsBuilder)
	if err != nil {
		return nil, err
	}

	network := &Network{
		tcpServer:      tcpServer,
		Websockets:     tcpServer.HttpServer.Websockets,
		KnownNodes:     knownNodes,
		BannedNodes:    bannedNodes,
		MempoolSync:    mempool_sync.NewMempoolSync(tcpServer.HttpServer.Websockets),
		KnownNodesSync: known_nodes_sync.NewNodesKnownSync(tcpServer.HttpServer.Websockets, knownNodes),
	}
	tcpServer.HttpServer.Initialize()

	network.continuouslyConnectNewPeers()

	network.continuouslyDownloadChain()

	if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
		network.continuouslyDownloadMempool()
		network.continuouslyDownloadNetworkNodes()
	}

	network.syncBlockchainNewConnections()

	return network, nil
}
