package websocks

import (
	"encoding/json"
	"errors"
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

	c, _, err := websocket.DefaultDialer.Dial(knownNode.UrlStr, nil)
	if err != nil {
		return
	}

	wsClient.conn = CreateAdvancedConnection(c, websockets.api, websockets.apiWebsockets)
	if err = websockets.NewConnection(wsClient.conn, true); err != nil {
		return
	}

	handshake := &api.APIHandshake{config.NAME, config.VERSION, string(config.NETWORK_SELECTED)}
	out := wsClient.conn.SendAwaitAnswer([]byte("handshake"), handshake)

	if out == nil {
		wsClient.Close()
		return nil, errors.New("Handshake was not received")
	}

	if out.err != nil {
		wsClient.Close()
		return nil, out.err
	}
	handshakeServer := new(api.APIHandshake)

	if err = json.Unmarshal(out.out, &handshakeServer); err != nil {
		wsClient.Close()
		return nil, errors.New("Handshake received was invalid")
	}

	if err = websockets.apiWebsockets.ValidateHandshake(handshakeServer); err != nil {
		wsClient.Close()
		return nil, errors.New("Handshake is invalid")
	}

	return
}
