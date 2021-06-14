package websocks

import (
	"context"
	"nhooyr.io/websocket"
	"pandora-pay/config"
	"pandora-pay/network/known-nodes"
	"pandora-pay/network/websocks/connection"
)

type WebsocketClient struct {
	knownNode  *known_nodes.KnownNode
	conn       *connection.AdvancedConnection
	websockets *Websockets
}

func (wsClient *WebsocketClient) Close(reason string) error {
	return wsClient.conn.Close(reason)
}

func CreateWebsocketClient(websockets *Websockets, knownNode *known_nodes.KnownNode) (*WebsocketClient, error) {

	wsClient := &WebsocketClient{
		knownNode:  knownNode,
		websockets: websockets,
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_TIMEOUT)
	defer cancel()

	c, _, err := websocket.Dial(ctx, knownNode.UrlStr, nil)
	if err != nil {
		return nil, err
	}

	if wsClient.conn, err = websockets.NewConnection(c, knownNode.UrlStr, false); err != nil {
		return nil, err
	}

	return wsClient, nil
}
