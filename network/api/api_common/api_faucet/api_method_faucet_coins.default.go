//go:build !wasm
// +build !wasm

package api_faucet

import (
	"errors"
	"go.jolheiser.com/hcaptcha"
	"net/http"
)

func (api *Faucet) GetFaucetCoins(r *http.Request, args *APIFaucetCoinsRequest, reply *APIFaucetCoinsReply) error {

	resp, err := api.hcpatchaClient.Verify(args.FaucetToken, hcaptcha.PostOptions{})
	if err != nil {
		return err
	}

	if !resp.Success {
		return errors.New("Faucet token is invalid")
	}

	//addr, err := api.wallet.GetWalletAddress(0, true)
	//if err != nil {
	//	return err
	//}

	//txData := &txs_builder.TxBuilderCreateZetherTxData{
	//	Payloads: []*txs_builder.TxBuilderCreateZetherTxPayload{{
	//		Sender:            addr.AddressEncoded,
	//		Asset:             config_coins.NATIVE_ASSET_FULL,
	//		Recipient:         args.Address,
	//		Data:              &wizard.WizardTransactionData{[]byte("Testnet Faucet Tx"), true},
	//		Fee:               &wizard.WizardZetherTransactionFee{&wizard.WizardTransactionFee{0, 0, 0, true}, false, 0, 0},
	//		Amount:            config.FAUCET_TESTNET_COINS_UNITS,
	//		RingConfiguration: &txs_builder.ZetherRingConfiguration{128, &txs_builder.ZetherSenderRingType{}, &txs_builder.ZetherRecipientRingType{}},
	//	}},
	//}
	//
	//ctx, cancel := context.WithCancel(context.Background())
	//defer cancel()
	//
	//tx, err := api.txsBuilder.CreateZetherTx(txData, nil, true, true, true, false, ctx, func(status string) {})
	//if err != nil {
	//	return err
	//}
	//
	//reply.Hash = tx.Bloom.Hash
	return nil

}
