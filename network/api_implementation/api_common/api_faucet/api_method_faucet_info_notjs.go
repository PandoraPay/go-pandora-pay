//go:build !wasm
// +build !wasm

package api_faucet

import (
	"net/http"
	"pandora-pay/network/network_config"
)

func (api *Faucet) GetFaucetInfo(r *http.Request, args *struct{}, reply *APIFaucetInfo) error {

	reply.FaucetTestnetEnabled = network_config.FAUCET_TESTNET_ENABLED
	if network_config.FAUCET_TESTNET_ENABLED {
		reply.Origin = network_config.NETWORK_ADDRESS_URL_STRING
		reply.ChallengeUri = "/static/challenge/challenge.html"
	}
	reply.FaucetTestnetCoins = network_config.FAUCET_TESTNET_COINS
	return nil
}
