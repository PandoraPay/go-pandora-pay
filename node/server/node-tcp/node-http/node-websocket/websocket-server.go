package node_websocket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"pandora-pay/gui"
)

type WebsocketServer struct {
}

var upgrader = websocket.Upgrader{}

func (ws *WebsocketServer) handleUpgradeConnection(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		gui.Error("ws error upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			gui.Info("ws error reading", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			gui.Info("write:", err)
			break
		}
	}
}

func (ws *WebsocketServer) InitializeWebsocketServer() {
	http.HandleFunc("/ws", ws.handleUpgradeConnection)
}

func CreateWebsocketServer() *WebsocketServer {
	return &WebsocketServer{}
}
