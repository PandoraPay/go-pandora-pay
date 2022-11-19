//go:build !wasm
// +build !wasm

package api_faucet

import (
	"go.jolheiser.com/hcaptcha"
	"pandora-pay/blockchain"
	"pandora-pay/mempool"
	"pandora-pay/network/network_config"
	"pandora-pay/wallet"
)

type Faucet struct {
	mempool        *mempool.Mempool
	chain          *blockchain.Blockchain
	wallet         *wallet.Wallet
	hcpatchaClient *hcaptcha.Client
}

func NewFaucet(mempool *mempool.Mempool, chain *blockchain.Blockchain, wallet *wallet.Wallet) (*Faucet, error) {

	api := &Faucet{
		mempool, chain, wallet, nil,
	}

	if network_config.FAUCET_TESTNET_ENABLED {
		// Dummy secret https://docs.hcaptcha.com/#integrationtest
		hcpatchaClient, err := hcaptcha.New(network_config.HCAPTCHA_SECRET_KEY)
		if err != nil {
			return nil, err
		}

		api.hcpatchaClient = hcpatchaClient
	}

	return api, nil
}
