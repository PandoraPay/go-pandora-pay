package client

import (
	"github.com/gorilla/websocket"
	"log"
	"net/url"
)

type WebsocketClient struct {
	URL  url.URL
	conn *websocket.Conn
}

func (ws *WebsocketClient) Close() error {
	return ws.conn.Close()
}

func (ws *WebsocketClient) Execute() {

	for {
		_, message, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s", message)
	}

}

func CreateWebsocketClient(url url.URL) (ws *WebsocketClient, err error) {

	ws = &WebsocketClient{
		URL: url,
	}

	ws.conn, _, err = websocket.DefaultDialer.Dial(ws.URL.String(), nil)
	if err != nil {
		return
	}

	return
}
