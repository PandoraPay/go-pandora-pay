//go:build !wasm
// +build !wasm

package api_faucet

import (
	"go.jolheiser.com/hcaptcha"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/transactions_builder"
	"pandora-pay/wallet"
)

type APICommonFaucet struct {
	mempool             *mempool.Mempool
	chain               *blockchain.Blockchain
	wallet              *wallet.Wallet
	transactionsBuilder *transactions_builder.TransactionsBuilder
	hcpatchaClient      *hcaptcha.Client
}

func CreateAPICommonFaucet(mempool *mempool.Mempool, chain *blockchain.Blockchain, wallet *wallet.Wallet, transactionsBuilder *transactions_builder.TransactionsBuilder) (*APICommonFaucet, error) {

	api := &APICommonFaucet{
		mempool, chain, wallet, transactionsBuilder, nil,
	}

	if config.FAUCET_TESTNET_ENABLED {
		// Dummy secret https://docs.hcaptcha.com/#integrationtest
		hcpatchaClient, err := hcaptcha.New(config.HCAPTCHA_SECRET_KEY)
		if err != nil {
			return nil, err
		}

		api.hcpatchaClient = hcpatchaClient
	}

	return api, nil
}
