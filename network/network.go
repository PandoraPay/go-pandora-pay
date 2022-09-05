package network

import (
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/known_nodes_sync"
	"pandora-pay/network/mempool_sync"
	"pandora-pay/network/server/node_tcp"
	"pandora-pay/network/websocks"
	"pandora-pay/settings"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_validator"
	"pandora-pay/wallet"
)

type Network struct {
	tcpServer      *node_tcp.TcpServer
	Websockets     *websocks.Websockets
	ConnectedNodes *connected_nodes.ConnectedNodes
	KnownNodes     *known_nodes.KnownNodes
	BannedNodes    *banned_nodes.BannedNodes
	MempoolSync    *mempool_sync.MempoolSync
	KnownNodesSync *known_nodes_sync.KnownNodesSync
}

func NewNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet, txsValidator *txs_validator.TxsValidator, txsBuilder *txs_builder.TxsBuilder) (*Network, error) {

	connectedNodes := connected_nodes.NewConnectedNodes()
	bannedNodes := banned_nodes.NewBannedNodes()

	knownNodes := known_nodes.NewKnownNodes(connectedNodes, bannedNodes)
	for _, seed := range config.NETWORK_SELECTED_SEEDS {
		knownNodes.AddKnownNode(seed.Url, true)
	}

	tcpServer, err := node_tcp.NewTcpServer(connectedNodes, bannedNodes, knownNodes, settings, chain, mempool, wallet, txsValidator, txsBuilder)
	if err != nil {
		return nil, err
	}

	network := &Network{
		tcpServer:      tcpServer,
		Websockets:     tcpServer.HttpServer.Websockets,
		ConnectedNodes: connectedNodes,
		KnownNodes:     knownNodes,
		BannedNodes:    bannedNodes,
		MempoolSync:    mempool_sync.NewMempoolSync(tcpServer.HttpServer.Websockets),
		KnownNodesSync: known_nodes_sync.NewNodesKnownSync(tcpServer.HttpServer.Websockets, knownNodes),
	}

	network.continuouslyConnectingNewPeers()

	network.continuouslyDownloadChain()

	if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
		network.continuouslyDownloadMempool()
		network.continuouslyDownloadNetworkNodes()
	}

	network.syncBlockchainNewConnections()

	return network, nil
}
