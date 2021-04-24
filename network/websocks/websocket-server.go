package websocks

import (
	"github.com/gorilla/websocket"
	"net/http"
	"pandora-pay/gui"
	"pandora-pay/network/websocks/connection"
)

type WebsocketServer struct {
	upgrader   websocket.Upgrader
	websockets *Websockets
}

func (wserver *WebsocketServer) handleUpgradeConnection(w http.ResponseWriter, r *http.Request) {

	c, err := wserver.upgrader.Upgrade(w, r, nil)
	if err != nil {
		gui.Error("ws error upgrade:", err)
		return
	}

	conn := connection.CreateAdvancedConnection(c, wserver.websockets.apiWebsockets.GetMap, true)
	if err = wserver.websockets.NewConnection(conn); err != nil {
		return
	}

}

func CreateWebsocketServer(websockets *Websockets) *WebsocketServer {

	wserver := &WebsocketServer{
		upgrader:   websocket.Upgrader{},
		websockets: websockets,
	}

	http.HandleFunc("/ws", wserver.handleUpgradeConnection)

	return wserver
}
