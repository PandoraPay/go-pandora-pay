//go:build wasm
// +build wasm

package api_faucet

import "net/http"

func (api *Faucet) GetFaucetCoins(r *http.Request, args *APIFaucetCoinsRequest, reply *APIFaucetCoinsReply) error {
	return nil
}
