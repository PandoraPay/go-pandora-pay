package websocks

import (
	"pandora-pay/network/known_nodes/known_node"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/websock"
)

type WebsocketClient struct {
	knownNode *known_node.KnownNodeScored
	conn      *connection.AdvancedConnection
}

func (this *websocketsType) NewWebsocketClient(knownNode *known_node.KnownNodeScored) (*WebsocketClient, error) {

	wsClient := &WebsocketClient{
		knownNode, nil,
	}

	c, err := websock.Dial(knownNode.URL)
	if err != nil {
		return nil, err
	}

	if wsClient.conn, err = this.NewConnection(c, knownNode.URL, knownNode, false); err != nil {
		return nil, err
	}

	return wsClient, nil
}
