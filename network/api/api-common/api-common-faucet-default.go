// +build !wasm

package api_common

import (
	"encoding/json"
	"errors"
	"go.jolheiser.com/hcaptcha"
	"pandora-pay/config"
	"pandora-pay/network/api/api-common/api_types"
)

type APICommonFaucet struct {
	hcpatchaClient *hcaptcha.Client
}

func (api *APICommonFaucet) GetFaucetInfo() ([]byte, error) {
	return json.Marshal(&api_types.APIFaucetInfo{
		config.HCAPTCHA_SITE_KEY,
		config.FAUCET_TESTNET_ENABLED,
		config.FAUCET_TESTNET_COINS,
	})
}

func (api *APICommonFaucet) GetFaucetCoins(request *api_types.APIFaucetCoinsRequest) ([]byte, error) {

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

	//publicKeyHash, err := request.GetPublicKeyHash()
	//if err != nil {
	//	return nil, err
	//}
	//
	//tx, err := api.transactionsBuilder.CreateSimpleTx([]string{testnet.wallet.Addresses[0].AddressEncoded}, 0, []uint64{testnet.nodes * config_stake.GetRequiredStake(blockHeight)}, [][]byte{config.NATIVE_TOKEN}, dsts, dstsAmounts, dstsTokens, 0, []byte{})
	//if err != nil {
	//	return
	//}

	return nil, nil
}

func createAPICommonFaucet() (*APICommonFaucet, error) {

	api := &APICommonFaucet{}

	// Dummy secret https://docs.hcaptcha.com/#integrationtest
	hcpatchaClient, err := hcaptcha.New(config.HCAPTCHA_SECRET_KEY)
	if err != nil {
		return nil, err
	}

	api.hcpatchaClient = hcpatchaClient

	return api, nil
}
