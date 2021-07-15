// +build wasm

package api_common

import (
	"pandora-pay/network/api/api-common/api_types"
	transactions_builder "pandora-pay/transactions-builder"
)

type APICommonFaucet struct {
}

func (api *APICommonFaucet) GetFaucetInfo() ([]byte, error) {
	return nil, nil
}

func (api *APICommonFaucet) GetFaucetCoins(request *api_types.APIFaucetCoinsRequest) ([]byte, error) {
	return nil, nil
}

func createAPICommonFaucet(wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*APICommonFaucet, error) {
	return &APICommonFaucet{}, nil
}
