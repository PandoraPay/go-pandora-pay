//go:build wasm
// +build wasm

package api_faucet

import (
	"net/url"
	"pandora-pay/network/websocks/connection"
)

func (api *Faucet) GetFaucetInfo_http(values *url.Values) (interface{}, error) {
	return nil, nil
}

func (api *Faucet) GetFaucetInfo_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return nil, nil
}
