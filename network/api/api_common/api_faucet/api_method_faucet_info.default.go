//go:build !wasm
// +build !wasm

package api_faucet

import (
	"net/http"
	"pandora-pay/config"
)

func (api *Faucet) GetFaucetInfo(r *http.Request, args *struct{}, reply *APIFaucetInfo) error {
	reply.HCaptchaSiteKey = config.HCAPTCHA_SITE_KEY
	reply.FaucetTestnetEnabled = config.FAUCET_TESTNET_ENABLED
	reply.FaucetTestnetCoins = config.FAUCET_TESTNET_COINS
	return nil
}
