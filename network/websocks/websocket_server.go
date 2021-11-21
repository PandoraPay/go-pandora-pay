package websocks

import (
	"net/http"
	"nhooyr.io/websocket"
	"pandora-pay/config"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/websocks/connection"
	"sync/atomic"
)

type WebsocketServer struct {
	websockets *Websockets
	knownNodes *known_nodes.KnownNodes
}

func (wserver *WebsocketServer) handleUpgradeConnection(w http.ResponseWriter, r *http.Request) {

	if atomic.LoadInt64(&wserver.websockets.serverSockets) >= config.WEBSOCKETS_NETWORK_SERVER_MAX {
		http.Error(w, "Too many websockets", 400)
		return
	}

	var err error

	var c *websocket.Conn
	var conn *connection.AdvancedConnection

	if c, err = websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true}); err != nil {
		//http.Error is not required because websocket.Accept will automatically send the error to the socket!
		return
	}

	if conn, err = wserver.websockets.NewConnection(c, r.RemoteAddr, nil, true); err != nil {
		return
	}

	if conn.Handshake.URL != "" {
		conn.KnownNode = wserver.knownNodes.AddKnownNode(conn.Handshake.URL, false)
	}

}

func CreateWebsocketServer(websockets *Websockets, knownNodes *known_nodes.KnownNodes) *WebsocketServer {

	wserver := &WebsocketServer{
		websockets: websockets,
		knownNodes: knownNodes,
	}

	http.HandleFunc("/ws", wserver.handleUpgradeConnection)

	return wserver
}
