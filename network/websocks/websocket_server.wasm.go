//go:build wasm
// +build wasm

package websocks

import "pandora-pay/network/known_nodes"

type WebsocketServer struct {
}

func CreateWebsocketServer(websockets *Websockets, knownNodes *known_nodes.KnownNodes) *WebsocketServer {
	return &WebsocketServer{}
}
