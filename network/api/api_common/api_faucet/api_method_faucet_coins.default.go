//go:build !wasm
// +build !wasm

package api_faucet

import (
	"context"
	"errors"
	"go.jolheiser.com/hcaptcha"
	"net/http"
	"pandora-pay/config/config_coins"
	"pandora-pay/txs_builder"
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

	txData := &txs_builder.TxBuilderCreateSimpleTx{
		0,
		&wizard.WizardTransactionData{[]byte("Testnet Faucet Tx"), false},
		&wizard.WizardTransactionFee{0, 0, 0, true},
		nil,
		[]*txs_builder.TxBuilderCreateSimpleTxVin{{
			addr.AddressEncoded,
			config_coins.ConvertToUnitsUint64Forced(100),
			config_coins.NATIVE_ASSET_FULL,
		}},
		[]*txs_builder.TxBuilderCreateSimpleTxVout{{
			args.Address,
			config_coins.ConvertToUnitsUint64Forced(100),
			config_coins.NATIVE_ASSET_FULL,
		}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tx, err := api.txsBuilder.CreateSimpleTx(txData, true, true, true, false, ctx, func(status string) {})
	if err != nil {
		return err
	}

	reply.Hash = tx.Bloom.Hash
	return nil
}
