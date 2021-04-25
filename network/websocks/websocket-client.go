package websocks

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/websocket"
	"pandora-pay/config"
	api_websockets "pandora-pay/network/api/api-websockets"
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

	wsClient.conn = connection.CreateAdvancedConnection(c, websockets.apiWebsockets.GetMap, false)
	if err = websockets.NewConnection(wsClient.conn); err != nil {
		return
	}

	handshake := &api_websockets.APIHandshake{config.NAME, config.VERSION, string(config.NETWORK_SELECTED)}
	handshakeBinary, _ := json.Marshal(handshake)

	out := wsClient.conn.SendAwaitAnswer([]byte("handshake"), handshakeBinary)

	if out.Err != nil {
		wsClient.Close()
		return nil, err
	}

	if out.Out == nil {
		wsClient.Close()
		return nil, errors.New("Handshake was not received")
	}

	handshakeServer := new(api_websockets.APIHandshake)
	if err = json.Unmarshal(out.Out, &handshakeServer); err != nil {
		wsClient.Close()
		return nil, errors.New("Handshake received was invalid")
	}

	if err = websockets.apiWebsockets.ValidateHandshake(handshakeServer); err != nil {
		wsClient.Close()
		return nil, errors.New("Handshake is invalid")
	}

	wsClient.conn.Send([]byte("chain-get"), nil)

	return
}
