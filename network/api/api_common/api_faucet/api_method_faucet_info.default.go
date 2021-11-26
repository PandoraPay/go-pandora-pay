//go:build !wasm
// +build !wasm

package api_faucet

import (
	"encoding/json"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/network/websocks/connection"
)

type APIFaucetInfo struct {
	HCaptchaSiteKey      string `json:"hCaptchaSiteKey,omitempty"`
	FaucetTestnetEnabled bool   `json:"faucetTestnetEnabled,omitempty"`
	FaucetTestnetCoins   uint64 `json:"faucetTestnetCoins,omitempty"`
}

func (api *Faucet) getFaucetInfo() ([]byte, error) {
	return json.Marshal(&APIFaucetInfo{
		config.HCAPTCHA_SITE_KEY,
		config.FAUCET_TESTNET_ENABLED,
		config.FAUCET_TESTNET_COINS,
	})
}

func (api *Faucet) GetFaucetInfo_http(values *url.Values) (interface{}, error) {
	return api.getFaucetInfo()
}

func (api *Faucet) GetFaucetInfo_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.getFaucetInfo()
}
