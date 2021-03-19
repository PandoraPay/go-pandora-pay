package websocks

import (
	"github.com/gorilla/websocket"
	"pandora-pay/config"
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

	handshake := &struct {
		Name    string
		Version string
		Network uint64
	}{config.NAME, config.VERSION, config.NETWORK_SELECTED}

	wsClient.conn.SendAwaitAnswer([]byte("handshake"), handshake)

	return
}
