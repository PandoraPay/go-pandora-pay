//go:build wasm
// +build wasm

package api_faucet

import (
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
)

type APICommonFaucet struct {
}

func (api *APICommonFaucet) GetFaucetInfo() ([]byte, error) {
	return nil, nil
}

func (api *APICommonFaucet) GetFaucetCoins(request *APIFaucetCoinsRequest) ([]byte, error) {
	return nil, nil
}

func CreateAPICommonFaucet(mempool *mempool.Mempool, chain *blockchain.Blockchain, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*APICommonFaucet, error) {
	return &APICommonFaucet{}, nil
}
