//go:build !wasm
// +build !wasm

package api_faucet

import (
	"context"
	"encoding/json"
	"errors"
	"go.jolheiser.com/hcaptcha"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/transactions_builder"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/wallet"
)

type APICommonFaucet struct {
	mempool             *mempool.Mempool
	chain               *blockchain.Blockchain
	wallet              *wallet.Wallet
	transactionsBuilder *transactions_builder.TransactionsBuilder
	hcpatchaClient      *hcaptcha.Client
}

func (api *APICommonFaucet) GetFaucetInfo() ([]byte, error) {
	return json.Marshal(&api_types.APIFaucetInfo{
		config.HCAPTCHA_SITE_KEY,
		config.FAUCET_TESTNET_ENABLED,
		config.FAUCET_TESTNET_COINS,
	})
}

func (api *APICommonFaucet) GetFaucetCoins(request *APIFaucetCoinsRequest) ([]byte, error) {

	if !config.FAUCET_TESTNET_ENABLED {
		return nil, errors.New("Faucet Testnet is not enabled")
	}

	resp, err := api.hcpatchaClient.Verify(request.FaucetToken, hcaptcha.PostOptions{})
	if err != nil {
		return nil, err
	}

	if !resp.Success {
		return nil, errors.New("Faucet token is invalid")
	}

	addr, err := api.wallet.GetWalletAddress(0)
	if err != nil {
		return nil, err
	}

	data := &wizard.TransactionsWizardData{[]byte("Testnet Faucet Tx"), false}
	fee := &wizard.TransactionsWizardFee{0, 0, 0, true}

	ringMembers, err := api.transactionsBuilder.CreateZetherRing(addr.AddressEncoded, request.Address, config_coins.NATIVE_ASSET_FULL, -1, -1)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tx, err := api.transactionsBuilder.CreateZetherTx([]string{addr.AddressEncoded}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{config.FAUCET_TESTNET_COINS_UNITS}, []string{request.Address}, []uint64{0}, [][]string{ringMembers}, []*wizard.TransactionsWizardData{data}, []*wizard.TransactionsWizardFee{fee}, true, false, false, false, ctx, func(status string) {})
	if err != nil {
		return nil, err
	}

	return tx.Bloom.Hash, nil

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
