package websockets

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"pandora-pay/gui"
)

type WebsocketServer struct {
	upgrader websocket.Upgrader
	sockets  *Websockets
}

func (ws *WebsocketServer) handleUpgradeConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		gui.Error("ws error upgrade:", err)
		return
	}

	if err = ws.sockets.NewConnection(conn, false); err != nil {
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

func CreateWebsocketServer(sockets *Websockets) *WebsocketServer {
	return &WebsocketServer{
		upgrader: websocket.Upgrader{},
		sockets:  sockets,
	}
}
