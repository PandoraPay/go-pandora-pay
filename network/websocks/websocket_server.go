package websocks

import (
	"net/http"
	"nhooyr.io/websocket"
	"pandora-pay/config"
	"sync/atomic"
)

type WebsocketServer struct {
	websockets *Websockets
}

func (wserver *WebsocketServer) handleUpgradeConnection(w http.ResponseWriter, r *http.Request) {

	if atomic.LoadInt64(&wserver.websockets.serverSockets) >= config.WEBSOCKETS_NETWORK_SERVER_MAX {
		http.Error(w, "Too many websockets", 400)
		return
	}

	var err error

	var c *websocket.Conn
	if c, err = websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true}); err != nil {
		//http.Error is not required because websocket.Accept will automatically send the error to the socket!
		return
	}

	if _, err = wserver.websockets.NewConnection(c, r.RemoteAddr, nil, true); err != nil {
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
