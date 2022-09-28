//go:build !wasm
// +build !wasm

package api_faucet

import (
	"net/http"
	"pandora-pay/config"
)

func (api *Faucet) GetFaucetInfo(r *http.Request, args *struct{}, reply *APIFaucetInfo) error {

	reply.FaucetTestnetEnabled = config.FAUCET_TESTNET_ENABLED
	if config.FAUCET_TESTNET_ENABLED {
		reply.Origin = config.NETWORK_ADDRESS_URL_STRING
		reply.ChallengeUri = "/static/challenge/challenge.html"
	}
	reply.FaucetTestnetCoins = config.FAUCET_TESTNET_COINS
	return nil
}
