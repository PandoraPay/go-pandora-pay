package websocks

import (
	"pandora-pay/network/known_nodes/known_node"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/websock"
)

type WebsocketClient struct {
	knownNode  *known_node.KnownNodeScored
	conn       *connection.AdvancedConnection
	websockets *Websockets
}

func NewWebsocketClient(websockets *Websockets, knownNode *known_node.KnownNodeScored) (*WebsocketClient, error) {

	wsClient := &WebsocketClient{
		knownNode:  knownNode,
		websockets: websockets,
	}

	c, err := websock.Dial(knownNode.URL)
	if err != nil {
		return nil, err
	}

	if wsClient.conn, err = websockets.NewConnection(c, knownNode.URL, knownNode, false); err != nil {
		return nil, err
	}

	return wsClient, nil
}
