//go:build js
// +build js

package websocks

import (
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes"
)

type WebsocketServer struct {
}

func NewWebsocketServer(websockets *Websockets, connectedNodes *connected_nodes.ConnectedNodes, knownNodes *known_nodes.KnownNodes) *WebsocketServer {
	return &WebsocketServer{}
}
