//go:build wasm
// +build wasm

package api_faucet

import "net/http"

func (api *Faucet) GetFaucetInfo(r *http.Request, args *struct{}, reply *APIFaucetInfo) error {
	return nil
}
