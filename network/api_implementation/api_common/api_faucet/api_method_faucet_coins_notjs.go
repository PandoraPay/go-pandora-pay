//go:build !wasm
// +build !wasm

package api_faucet

import (
	"context"
	"errors"
	"go.jolheiser.com/hcaptcha"
	"net/http"
	"pandora-pay/config"
	"pandora-pay/config/config_coins"
	"pandora-pay/txs_builder"
	"pandora-pay/txs_builder/txs_builder_zether_helper"
	"pandora-pay/txs_builder/wizard"
)

func (api *Faucet) GetFaucetCoins(r *http.Request, args *APIFaucetCoinsRequest, reply *APIFaucetCoinsReply) error {

	resp, err := api.hcpatchaClient.Verify(args.FaucetToken, hcaptcha.PostOptions{})
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New("Faucet token is invalid")
	}

	addr, err := api.wallet.GetWalletAddress(0, true)
	if err != nil {
		return err
	}

	txData := &txs_builder.TxBuilderCreateZetherTxData{
		Payloads: []*txs_builder.TxBuilderCreateZetherTxPayload{{
			txs_builder_zether_helper.TxsBuilderZetherTxPayloadBase{
				addr.AddressEncoded,
				args.Address,
				128,
				nil,
			},
			config_coins.NATIVE_ASSET_FULL,
			config.FAUCET_TESTNET_COINS_UNITS,
			0,
			&txs_builder.ZetherRingConfiguration{&txs_builder.ZetherSenderRingType{}, &txs_builder.ZetherRecipientRingType{}},
			0,
			&wizard.WizardTransactionData{[]byte("Testnet Faucet Tx"), true},
			&wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0},
			nil,
		}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tx, err := txs_builder.TxsBuilder.CreateZetherTx(txData, nil, true, true, true, false, ctx, func(status string) {})
	if err != nil {
		return err
	}

	reply.Hash = tx.Bloom.Hash
	return nil

}
