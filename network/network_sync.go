package network

import (
	"pandora-pay/gui"
	"pandora-pay/helpers/recovery"
	"pandora-pay/network/api/api_websockets/consensus"
	"pandora-pay/network/known_nodes/known_node"
	"pandora-pay/network/network_config"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection"
	"time"
)

func (network *Network) continuouslyConnectingNewPeers() {

	for i := 0; i < network_config.WEBSOCKETS_CONCURRENT_NEW_CONENCTIONS; i++ {
		index := i
		recovery.SafeGo(func() {

			for {

				if network.Websockets.GetClients() >= network_config.WEBSOCKETS_NETWORK_CLIENTS_MAX {
					time.Sleep(500 * time.Millisecond)
					continue
				}

				var knownNode *known_node.KnownNodeScored
				if index == 0 {
					knownNode = network.KnownNodes.GetBestNotConnectedKnownNode()
				} else {
					knownNode = network.KnownNodes.GetRandomKnownNode()
				}
				if knownNode != nil {

					//gui.GUI.Log("connecting to", knownNode.URL, atomic.LoadInt32(&knownNode.Score))

					if network.BannedNodes.IsBanned(knownNode.URL) {
						network.KnownNodes.DecreaseKnownNodeScore(knownNode, -10, false)
					} else {
						_, err := websocks.NewWebsocketClient(network.Websockets, knownNode)
						if err != nil {

							//gui.GUI.Error("error connecting", knownNode.URL, err)

							if err.Error() != "Already connected" {
								network.KnownNodes.DecreaseKnownNodeScore(knownNode, -20, false)
							}

						} else {
							gui.GUI.Log("connected to: " + knownNode.URL)
						}
					}
				}

				time.Sleep(100 * time.Millisecond)
			}
		})
	}

}

func (network *Network) continuouslyDownloadChain() {
	recovery.SafeGo(func() {

		for {

			if conn := network.Websockets.GetRandomSocket(); conn != nil {
				data, err := connection.SendJSONAwaitAnswer[consensus.ChainUpdateNotification](conn, []byte("get-chain"), nil, nil, 0)
				if err == nil {
					network.Websockets.ApiWebsockets.Consensus.ChainUpdateProcess(conn, data)
				}
			}

			time.Sleep(2000 * time.Millisecond)
		}

	})
}

func (network *Network) continuouslyDownloadMempool() {

	recovery.SafeGo(func() {

		for {

			if conn := network.Websockets.GetRandomSocket(); conn != nil {
				if network_config.CONSENSUS == network_config.CONSENSUS_TYPE_FULL && conn.Handshake.Consensus == network_config.CONSENSUS_TYPE_FULL {
					network.MempoolSync.DownloadMempool(conn)
				}
			}

			time.Sleep(2000 * time.Millisecond)
		}

	})

}

func (network *Network) continuouslyDownloadNetworkNodes() {

	recovery.SafeGo(func() {

		for {

			conn := network.Websockets.GetRandomSocket()
			if conn != nil {

				if network_config.CONSENSUS == network_config.CONSENSUS_TYPE_FULL && conn.Handshake.Consensus == network_config.CONSENSUS_TYPE_FULL {
					network.KnownNodesSync.DownloadNetworkNodes(conn)
				}

			}

			time.Sleep(10000 * time.Millisecond)
		}

	})

}

func (network *Network) syncBlockchainNewConnections() {
	recovery.SafeGo(func() {

		cn := network.Websockets.UpdateNewConnectionMulticast.AddListener()
		defer network.Websockets.UpdateNewConnectionMulticast.RemoveChannel(cn)

		for {

			conn, ok := <-cn
			if !ok {
				return
			}

			//making it async
			recovery.SafeGo(func() {

				data, err := connection.SendJSONAwaitAnswer[consensus.ChainUpdateNotification](conn, []byte("get-chain"), nil, nil, 0)
				if err == nil {
					network.Websockets.ApiWebsockets.Consensus.ChainUpdateProcess(conn, data)
				}

			})

		}
	})
}
