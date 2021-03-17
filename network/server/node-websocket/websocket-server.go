package node_websocket

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"pandora-pay/gui"
)

type WebsocketServer struct {
	upgrader websocket.Upgrader
}

func (ws *WebsocketServer) handleUpgradeConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		gui.Error("ws error upgrade:", err)
		return
	}
	defer conn.Close()
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			gui.Info("ws error reading", err)
			break
		}
		log.Printf("recv: %s", message)
		err = conn.WriteMessage(mt, message)
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
	return &WebsocketServer{
		upgrader: websocket.Upgrader{},
	}
}
