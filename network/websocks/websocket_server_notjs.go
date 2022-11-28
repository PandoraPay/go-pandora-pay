//go:build !js
// +build !js

package websocks

import (
	"net/http"
	"pandora-pay/helpers/recovery"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/network_config"
	"pandora-pay/network/websocks/websock"
	"sync/atomic"
)

func (this *websocketsType) HandleUpgradeConnection(w http.ResponseWriter, r *http.Request) {

	if atomic.LoadInt64(&connected_nodes.ConnectedNodes.ServerSockets) >= network_config.WEBSOCKETS_NETWORK_SERVER_MAX {
		http.Error(w, "Too many websockets", 400)
		return
	}

	c, err := websock.Upgrade(w, r)
	if err != nil {
		return
	}

	conn, err := Websockets.NewConnection(c, r.RemoteAddr, nil, true)
	if err != nil {
		return
	}

	if conn.Handshake.URL != "" {
		conn.KnownNode, err = known_nodes.KnownNodes.AddKnownNode(conn.Handshake.URL, false)
		if conn.KnownNode != nil {
			recovery.SafeGo(conn.IncreaseKnownNodeScore)
		}
	}

}
