package network

import (
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/consensus"
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
	KnownNodes     *known_nodes.KnownNodes
	BannedNodes    *banned_nodes.BannedNodes
	MempoolSync    *mempool_sync.MempoolSync
	KnownNodesSync *known_nodes_sync.KnownNodesSync
	Websockets     *websocks.Websockets
	Consensus      *consensus.Consensus
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

	network := &Network{
		tcpServer:      tcpServer,
		KnownNodes:     knownNodes,
		BannedNodes:    bannedNodes,
		MempoolSync:    mempool_sync.CreateMempoolSync(tcpServer.HttpServer.Websockets),
		KnownNodesSync: known_nodes_sync.CreateNodesKnownSync(tcpServer.HttpServer.Websockets, knownNodes),
		Websockets:     tcpServer.HttpServer.Websockets,
		Consensus:      consensus.CreateConsensus(tcpServer.HttpServer, chain, mempool),
	}
	tcpServer.HttpServer.Initialize()

	network.continuouslyConnectNewPeers()

	if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
		network.continuouslyDownloadMempool()
		network.continuouslyDownloadNetworkNodes()
	}

	network.syncBlockchainNewConnections()

	return network, nil
}
