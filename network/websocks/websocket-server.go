package websocks

import (
	"github.com/gorilla/websocket"
	"net/http"
	"pandora-pay/gui"
)

type WebsocketServer struct {
	upgrader websocket.Upgrader
	sockets  *Websockets
}

func (wserver *WebsocketServer) handleUpgradeConnection(w http.ResponseWriter, r *http.Request) {

	c, err := wserver.upgrader.Upgrade(w, r, nil)
	if err != nil {
		gui.Error("ws error upgrade:", err)
		return
	}

	conn := CreateAdvancedConnection(c, wserver.sockets.api, wserver.sockets.apiWebsockets)
	if err = wserver.sockets.NewConnection(conn, false); err != nil {
		return
	}

}

func CreateWebsocketServer(sockets *Websockets) *WebsocketServer {

	wserver := &WebsocketServer{
		upgrader: websocket.Upgrader{},
		sockets:  sockets,
	}

	http.HandleFunc("/ws", wserver.handleUpgradeConnection)

	return wserver
}
