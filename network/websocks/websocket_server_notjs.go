//go:build !js
// +build !js

package websocks

import (
	"net/http"
	"pandora-pay/config"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/websocks/websock"
	"pandora-pay/recovery"
	"sync/atomic"
)

type WebsocketServer struct {
	websockets     *Websockets
	connectedNodes *connected_nodes.ConnectedNodes
	knownNodes     *known_nodes.KnownNodes
}

func (wserver *WebsocketServer) HandleUpgradeConnection(w http.ResponseWriter, r *http.Request) {

	if atomic.LoadInt64(&wserver.connectedNodes.ServerSockets) >= config.WEBSOCKETS_NETWORK_SERVER_MAX {
		http.Error(w, "Too many websockets", 400)
		return
	}

	c, err := websock.Upgrade(w, r)
	if err != nil {
		return
	}

	conn, err := wserver.websockets.NewConnection(c, r.RemoteAddr, nil, true)
	if err != nil {
		return
	}

	if conn.Handshake.URL != "" {
		conn.KnownNode, err = wserver.knownNodes.AddKnownNode(conn.Handshake.URL, false)
		if conn.KnownNode != nil {
			recovery.SafeGo(conn.IncreaseKnownNodeScore)
		}
	}

}

func NewWebsocketServer(websockets *Websockets, connectedNodes *connected_nodes.ConnectedNodes, knownNodes *known_nodes.KnownNodes) *WebsocketServer {

	wserver := &WebsocketServer{
		websockets,
		connectedNodes,
		knownNodes,
	}

	return wserver
}
