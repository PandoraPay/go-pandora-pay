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

	wserver.upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	c, err := wserver.upgrader.Upgrade(w, r, nil)
	if err != nil {
		gui.Error("ws error upgrade:", err)
		return
	}

	conn := connection.CreateAdvancedConnection(c, wserver.websockets.ApiWebsockets.GetMap, true)
	if err = wserver.websockets.NewConnection(conn); err != nil {
		return
	}

	if err = wserver.websockets.InitializeConnection(conn); err != nil {
		return
	}

}

func CreateWebsocketServer(websockets *Websockets) *WebsocketServer {

	wserver := &WebsocketServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		websockets: websockets,
	}

	http.HandleFunc("/ws", wserver.handleUpgradeConnection)

	return wserver
}
