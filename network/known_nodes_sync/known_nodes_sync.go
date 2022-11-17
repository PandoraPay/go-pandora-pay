package known_nodes_sync

import (
	"pandora-pay/network/api_implementation/api_common"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection"
)

type KnownNodesSync struct {
	websockets *websocks.Websockets
	knownNodes *known_nodes.KnownNodes
}

func (self *KnownNodesSync) DownloadNetworkNodes(conn *connection.AdvancedConnection) error {

	data, err := connection.SendJSONAwaitAnswer[api_common.APINetworkNodesReply](conn, []byte("network/nodes"), nil, nil, 0)
	if err != nil {
		return err
	}

	if data == nil || data.Nodes == nil {
		return nil
	}

	for _, node := range data.Nodes {
		self.knownNodes.AddKnownNode(node.URL, false)
	}

	return nil
}

func NewNodesKnownSync(websockets *websocks.Websockets, knownNodes *known_nodes.KnownNodes) *KnownNodesSync {
	return &KnownNodesSync{
		websockets: websockets,
		knownNodes: knownNodes,
	}
}
