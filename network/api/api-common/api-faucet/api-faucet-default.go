//go:build !wasm
// +build !wasm

package api_faucet

import (
	"encoding/json"
	"errors"
	"go.jolheiser.com/hcaptcha"
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api-common/api_types"
	transactions_builder "pandora-pay/transactions-builder"
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

	panic("err")

	//addr, err := api.wallet.GetWalletAddress(0)
	//if err != nil {
	//	return nil, err
	//}

	//tx, err := api.transactionsBuilder.CreateZetherTx([]string{addr.AddressEncoded}, 0, config.NATIVE_TOKEN_FULL, []uint64{config.FAUCET_TESTNET_COINS_UNITS}, []string{request.Address}, []uint64{config.FAUCET_TESTNET_COINS_UNITS}, &wizard.TransactionsWizardData{[]byte("Testnet Faucet Tx"), false}, &wizard.TransactionsWizardFee{0, 0, true}, true, false, false, func(status string) {
	//
	//})
	//if err != nil {
	//	return nil, err
	//}

	//return tx.Bloom.Hash, nil

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
