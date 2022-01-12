//go:build !wasm
// +build !wasm

package api_faucet

import (
	"net/http"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/network/websocks/connection"
)

func (api *Faucet) FaucetInfo(r *http.Request, args *struct{}, reply *APIFaucetInfo) error {
	reply.HCaptchaSiteKey = config.HCAPTCHA_SITE_KEY
	reply.FaucetTestnetEnabled = config.FAUCET_TESTNET_ENABLED
	reply.FaucetTestnetCoins = config.FAUCET_TESTNET_COINS
	return nil
}

func (api *Faucet) GetFaucetInfo_http(values url.Values) (interface{}, error) {
	reply := &APIFaucetInfo{}
	return reply, api.FaucetInfo(nil, nil, reply)
}

func (api *Faucet) GetFaucetInfo_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &APIFaucetInfo{}
	return reply, api.FaucetInfo(nil, nil, reply)
}
