package websocks

import (
	"context"
	"nhooyr.io/websocket"
	"pandora-pay/config"
	"pandora-pay/network/known-nodes"
	"pandora-pay/network/websocks/connection"
)

type WebsocketClient struct {
	knownNode           *known_nodes.KnownNode
	conn                *connection.AdvancedConnection
	websockets          *Websockets
	handshakeValidation bool
}

func (wsClient *WebsocketClient) Close(reason string) error {
	return wsClient.conn.Close(reason)
}

func CreateWebsocketClient(websockets *Websockets, knownNode *known_nodes.KnownNode) (wsClient *WebsocketClient, err error) {

	wsClient = &WebsocketClient{
		knownNode:  knownNode,
		websockets: websockets,
	}

	ctx, cancel := context.WithTimeout(context.Background(), config.WEBSOCKETS_TIMEOUT)
	defer cancel()

	c, _, err := websocket.Dial(ctx, knownNode.UrlStr, nil)
	if err != nil {
		return
	}

	wsClient.conn = connection.CreateAdvancedConnection(c, knownNode.UrlStr, websockets.ApiWebsockets.GetMap, false)
	if err = websockets.NewConnection(wsClient.conn); err != nil {
		return
	}

	if err = websockets.InitializeConnection(wsClient.conn); err != nil {
		return
	}

	return
}
