package network

import (
	"context"
	"github.com/vmihailenco/msgpack/v5"
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
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/settings"
	"pandora-pay/wallet"
	"time"
)

type NetworkType struct {
	tcpServer      *node_tcp.TcpServer
	Websockets     *websocks.Websockets
	ConnectedNodes *connected_nodes.ConnectedNodes
	KnownNodes     *known_nodes.KnownNodes
	BannedNodes    *banned_nodes.BannedNodes
	MempoolSync    *mempool_sync.MempoolSync
	KnownNodesSync *known_nodes_sync.KnownNodesSync
}

var Network *NetworkType

func (network *NetworkType) Send(name, data []byte, ctxDuration time.Duration) error {

	for {

		<-network.Websockets.ReadyCn.Load()
		list := network.ConnectedNodes.AllList.Get()
		if len(list) > 0 {
			sock := list[0]
			if err := sock.Send(name, data, ctxDuration); err != nil {
				return err
			}
			return nil
		}
	}

}

func (network *NetworkType) SendJSON(name, data []byte, ctxDuration time.Duration) error {
	out, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}

	return network.Send(name, out, ctxDuration)
}

func (network *NetworkType) SendAwaitAnswer(name, data []byte, ctxParent context.Context, ctxDuration time.Duration) *advanced_connection_types.AdvancedConnectionReply {
	for {
		<-network.Websockets.ReadyCn.Load()
		list := network.ConnectedNodes.AllList.Get()
		if len(list) > 0 {
			sock := list[0]
			result := sock.SendAwaitAnswer(name, data, ctxParent, ctxDuration)
			if result.Timeout {
				continue
			}
			return result
		}
	}
}

func SendJSONAwaitAnswer[T any](name []byte, data any, ctxParent context.Context, ctxDuration time.Duration) (*T, error) {

	out, err := msgpack.Marshal(data)
	if err != nil {
		return nil, err
	}

	for {
		<-Network.Websockets.ReadyCn.Load()
		list := Network.ConnectedNodes.AllList.Get()
		if len(list) > 0 {
			sock := list[0]

			out := sock.SendAwaitAnswer(name, out, ctxParent, ctxDuration)
			if out.Err != nil {
				if out.Timeout {
					continue
				}
				return nil, out.Err
			}

			final := new(T)
			if err = msgpack.Unmarshal(out.Out, final); err != nil {
				return nil, err
			}
			return final, nil
		}
	}
}

func NewNetwork(settings *settings.Settings, chain *blockchain.Blockchain, mempool *mempool.Mempool, wallet *wallet.Wallet) error {

	connectedNodes := connected_nodes.NewConnectedNodes()
	bannedNodes := banned_nodes.NewBannedNodes()

	knownNodes := known_nodes.NewKnownNodes(connectedNodes, bannedNodes)

	list := make([]string, len(config.NETWORK_SELECTED_SEEDS))
	for i, seed := range config.NETWORK_SELECTED_SEEDS {
		list[i] = seed.Url
	}
	if err := knownNodes.Reset(list, true); err != nil {
		return err
	}

	tcpServer, err := node_tcp.NewTcpServer(connectedNodes, bannedNodes, knownNodes, settings, chain, mempool, wallet)
	if err != nil {
		return err
	}

	network := &NetworkType{
		tcpServer,
		tcpServer.HttpServer.Websockets,
		connectedNodes,
		knownNodes,
		bannedNodes,
		mempool_sync.NewMempoolSync(tcpServer.HttpServer.Websockets),
		known_nodes_sync.NewNodesKnownSync(tcpServer.HttpServer.Websockets, knownNodes),
	}

	network.continuouslyConnectingNewPeers()

	network.continuouslyDownloadChain()

	if config.NODE_CONSENSUS == config.NODE_CONSENSUS_TYPE_FULL {
		network.continuouslyDownloadMempool()
		network.continuouslyDownloadNetworkNodes()
	}

	network.syncBlockchainNewConnections()

	Network = network
	return nil
}
