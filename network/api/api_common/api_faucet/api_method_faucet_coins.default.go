//go:build !wasm
// +build !wasm

package api_faucet

import (
	"context"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"go.jolheiser.com/hcaptcha"
	"net/http"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_builder/wizard"
)

func (api *Faucet) FaucetCoins(r *http.Request, args *APIFaucetCoinsRequest, reply *helpers.HexBytes) error {

	if !config.FAUCET_TESTNET_ENABLED {
		return errors.New("Faucet Testnet is not enabled")
	}

	resp, err := api.hcpatchaClient.Verify(args.FaucetToken, hcaptcha.PostOptions{})
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New("Faucet token is invalid")
	}

	addr, err := api.wallet.GetWalletAddress(1, true)
	if err != nil {
		return err
	}

	data := &wizard.WizardTransactionData{[]byte("Testnet Faucet Tx"), false}
	fees := []*wizard.WizardZetherTransactionFee{{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0}}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tx, err := api.txsBuilder.CreateZetherTx([]wizard.WizardZetherPayloadExtra{nil}, []string{addr.AddressEncoded}, [][]byte{config_coins.NATIVE_ASSET_FULL}, []uint64{config.FAUCET_TESTNET_COINS_UNITS}, []string{args.Address}, []uint64{0}, []*txs_builder.ZetherRingConfiguration{{-1, -1}}, []*wizard.WizardTransactionData{data}, fees, true, true, true, false, ctx, func(status string) {})
	if err != nil {
		return err
	}

	*reply = tx.Bloom.Hash
	return nil

}

func (api *Faucet) GetFaucetCoins_http(values url.Values) (interface{}, error) {
	args := &APIFaucetCoinsRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	var reply helpers.HexBytes
	return reply, api.FaucetCoins(nil, args, &reply)
}

func (api *Faucet) GetFaucetCoins_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIFaucetCoinsRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	var reply helpers.HexBytes
	return reply, api.FaucetCoins(nil, args, &reply)
}
