package websocks

import (
	"net/http"
	"nhooyr.io/websocket"
	"pandora-pay/network/websocks/connection"
)

type WebsocketServer struct {
	websockets *Websockets
}

func (wserver *WebsocketServer) handleUpgradeConnection(w http.ResponseWriter, r *http.Request) {

	var err error

	var c *websocket.Conn
	if c, err = websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true}); err != nil {
		return
	}

	conn := connection.CreateAdvancedConnection(c, r.RemoteAddr, wserver.websockets.ApiWebsockets.GetMap, true)
	if err = wserver.websockets.NewConnection(conn); err != nil {
		return
	}

	if err = wserver.websockets.InitializeConnection(conn); err != nil {
		return
	}

}

func CreateWebsocketServer(websockets *Websockets) *WebsocketServer {

	wserver := &WebsocketServer{
		websockets: websockets,
	}

	http.HandleFunc("/ws", wserver.handleUpgradeConnection)

	return wserver
}
