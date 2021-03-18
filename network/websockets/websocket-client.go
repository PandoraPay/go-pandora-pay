package websockets

import (
	"github.com/gorilla/websocket"
	"log"
	"pandora-pay/network/known-nodes"
)

type WebsocketClient struct {
	knownNode           *known_nodes.KnownNode
	conn                *websocket.Conn
	socks               *Websockets
	handshakeValidation bool
}

func (wsClient *WebsocketClient) Close() error {
	return wsClient.conn.Close()
}

func (wsClient *WebsocketClient) Execute() {

	for {
		_, message, err := wsClient.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)
	}

}

func CreateWebsocketClient(socks *Websockets, knownNode *known_nodes.KnownNode) (wsClient *WebsocketClient, err error) {

	wsClient = &WebsocketClient{
		knownNode: knownNode,
		socks:     socks,
	}

	wsClient.conn, _, err = websocket.DefaultDialer.Dial(knownNode.Url.String(), nil)
	if err != nil {
		return
	}

	if err = socks.NewConnection(wsClient.conn, true); err != nil {
		return
	}

	wsClient.Execute()

	return
}
