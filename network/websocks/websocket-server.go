package websocks

import (
	"net/http"
	"nhooyr.io/websocket"
	"pandora-pay/config"
	"pandora-pay/network/websocks/connection"
	"sync/atomic"
)

type WebsocketServer struct {
	websockets *Websockets
}

func (wserver *WebsocketServer) handleUpgradeConnection(w http.ResponseWriter, r *http.Request) {

	if atomic.LoadInt64(&wserver.websockets.ServerSockets) >= config.WEBSOCKETS_NETWORK_SERVER_MAX {
		http.Error(w, "Too many websockets", 400)
		return
	}

	var err error

	var c *websocket.Conn
	if c, err = websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true}); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	conn := connection.CreateAdvancedConnection(c, r.RemoteAddr, wserver.websockets.ApiWebsockets.GetMap, true)
	if err = wserver.websockets.NewConnection(conn); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	if err = wserver.websockets.InitializeConnection(conn); err != nil {
		http.Error(w, err.Error(), 400)
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
