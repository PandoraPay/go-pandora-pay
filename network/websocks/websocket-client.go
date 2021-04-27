package websocks

import (
	"github.com/gorilla/websocket"
	"pandora-pay/network/known-nodes"
	"pandora-pay/network/websocks/connection"
)

type WebsocketClient struct {
	knownNode           *known_nodes.KnownNode
	conn                *connection.AdvancedConnection
	websockets          *Websockets
	handshakeValidation bool
}

func (wsClient *WebsocketClient) Close() error {
	return wsClient.conn.Close()
}

func CreateWebsocketClient(websockets *Websockets, knownNode *known_nodes.KnownNode) (wsClient *WebsocketClient, err error) {

	wsClient = &WebsocketClient{
		knownNode:  knownNode,
		websockets: websockets,
	}

	c, _, err := websocket.DefaultDialer.Dial(knownNode.UrlStr, nil)
	if err != nil {
		return
	}

	wsClient.conn = connection.CreateAdvancedConnection(c, websockets.ApiWebsockets.GetMap, false)
	if err = websockets.NewConnection(wsClient.conn); err != nil {
		return
	}

	if err = websockets.InitializeConnection(wsClient.conn); err != nil {
		return
	}

	return
}
