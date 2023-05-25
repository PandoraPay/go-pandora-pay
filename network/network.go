package network

import (
	"context"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/helpers/msgpack"
	"pandora-pay/mempool"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/server/node_tcp"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/settings"
	"pandora-pay/wallet"
	"time"
)

type networkType struct {
}

var Network *networkType

func (this *networkType) Send(name, data []byte, ctxDuration time.Duration) error {

	for {

		<-websocks.Websockets.ReadyCn.Load()
		list := connected_nodes.ConnectedNodes.AllList.Get()
		if len(list) > 0 {
			sock := list[0]
			if err := sock.Send(name, data, ctxDuration); err != nil {
				return err
			}
			return nil
		}
	}

}

func (this *networkType) SendJSON(name, data []byte, ctxDuration time.Duration) error {
	out, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}

	return this.Send(name, out, ctxDuration)
}

func (this *networkType) SendAwaitAnswer(name, data []byte, ctxParent context.Context, ctxDuration time.Duration) *advanced_connection_types.AdvancedConnectionReply {
	for {
		<-websocks.Websockets.ReadyCn.Load()
		list := connected_nodes.ConnectedNodes.AllList.Get()
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
		<-websocks.Websockets.ReadyCn.Load()
		list := connected_nodes.ConnectedNodes.AllList.Get()
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

	list := make([]string, len(config.NETWORK_SELECTED_SEEDS))
	for i, seed := range config.NETWORK_SELECTED_SEEDS {
		list[i] = seed.Url
	}
	if err := known_nodes.KnownNodes.Reset(list, true); err != nil {
		return err
	}

	if err := node_tcp.NewTcpServer(settings, chain, mempool, wallet); err != nil {
		return err
	}

	Network = &networkType{}

	Network.continuouslyConnectingNewPeers()
	Network.continuouslyDownloadNetworkNodes()

	return nil
}
