//go:build !wasm
// +build !wasm

package api_faucet

import (
	"context"
	"encoding/json"
	"errors"
	"go.jolheiser.com/hcaptcha"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/transactions_builder"
	"pandora-pay/transactions_builder/wizard"
)

func (api *Faucet) getFaucetCoins(request *APIFaucetCoinsRequest) ([]byte, error) {

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

	addr, err := api.wallet.GetWalletAddress(1)
	if err != nil {
		return nil, err
	}

	data := &wizard.WizardTransactionData{[]byte("Testnet Faucet Tx"), false}
	fees := []*wizard.WizardZetherTransactionFee{{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tx, err := api.transactionsBuilder.CreateZetherTx([]wizard.WizardZetherPayloadExtra{nil}, []string{addr.AddressEncoded}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{config.FAUCET_TESTNET_COINS_UNITS}, []string{request.Address}, []uint64{0}, []*transactions_builder.ZetherRingConfiguration{{-1, -1}}, []*wizard.WizardTransactionData{data}, fees, true, false, false, false, ctx, func(status string) {})
	if err != nil {
		return nil, err
	}

	return tx.Bloom.Hash, nil

}

func (api *Faucet) GetFaucetCoins_http(values *url.Values) (interface{}, error) {

	request := &APIFaucetCoinsRequest{"", ""}

	if values.Get("address") != "" {
		request.Address = values.Get("address")
	} else {
		return nil, errors.New("parameter 'address' was not specified")
	}

	if values.Get("faucetToken") != "" {
		request.FaucetToken = values.Get("faucetToken")
	}

	return api.getFaucetCoins(request)
}

func (api *Faucet) GetFaucetCoins_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIFaucetCoinsRequest{"", ""}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}
	return api.getFaucetCoins(request)
}
