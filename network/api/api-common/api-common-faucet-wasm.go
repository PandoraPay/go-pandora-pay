// +build wasm

package api_common

import (
	"pandora-pay/network/api/api-common/api_types"
)

type APICommonFaucet struct {
}

func (api *APICommonFaucet) GetFaucetInfo() ([]byte, error) {
	return nil, nil
}

func (api *APICommonFaucet) GetFaucetCoins(request *api_types.APIFaucetCoinsRequest) ([]byte, error) {
	return nil, nil
}

func createAPICommonFaucet() (*APICommonFaucet, error) {
	return &APICommonFaucet{}, nil
}
