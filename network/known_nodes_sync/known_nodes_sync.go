package known_nodes_sync

import (
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/websocks"
	"pandora-pay/network/websocks/connection"
)

type KnownNodesSync struct {
	websockets *websocks.Websockets
	knownNodes *known_nodes.KnownNodes
}

func (self *KnownNodesSync) DownloadNetworkNodes(conn *connection.AdvancedConnection) (err error) {

	out := conn.SendJSONAwaitAnswer([]byte("network/nodes"), nil, nil)
	if out.Err != nil {
		return
	}

	data := &api_common.APINetworkNodesReply{}
	if err = msgpack.Unmarshal(out.Out, data); err != nil {
		return
	}

	for _, node := range data.Nodes {
		self.knownNodes.AddKnownNode(node.URL, false)
	}

	return
}

func NewNodesKnownSync(websockets *websocks.Websockets, knownNodes *known_nodes.KnownNodes) *KnownNodesSync {
	return &KnownNodesSync{
		websockets: websockets,
		knownNodes: knownNodes,
	}
}
