package websocks

import (
	"github.com/gorilla/websocket"
	"pandora-pay/config"
	"pandora-pay/network/api"
	"pandora-pay/network/known-nodes"
)

type WebsocketClient struct {
	knownNode           *known_nodes.KnownNode
	conn                *AdvancedConnection
	websockets          *Websockets
	handshakeValidation bool
}

func (wsClient *WebsocketClient) Close() error {
	return wsClient.conn.Conn.Close()
}

func CreateWebsocketClient(websockets *Websockets, knownNode *known_nodes.KnownNode) (wsClient *WebsocketClient, err error) {

	wsClient = &WebsocketClient{
		knownNode:  knownNode,
		websockets: websockets,
	}

	c, _, err := websocket.DefaultDialer.Dial(knownNode.Url.String(), nil)
	if err != nil {
		return
	}

	wsClient.conn = CreateAdvancedConnection(c, websockets.api, websockets.apiWebsockets)
	if err = websockets.NewConnection(wsClient.conn, true); err != nil {
		return
	}

	handshake := api.APIHandshake{
		Name:    config.NAME,
		Version: config.VERSION,
		Network: config.NETWORK_SELECTED,
	}
	wsClient.conn.SendAwaitAnswer([]byte("handshake"), handshake)

	return
}
