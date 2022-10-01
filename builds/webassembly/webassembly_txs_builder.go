package main

import (
	"encoding/json"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/app"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple/transaction_simple_extra"
	"pandora-pay/builds/webassembly/webassembly_utils"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/txs_builder/wizard"
	"syscall/js"
)

func createSimpleTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeObject || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		txData := &struct {
			TxScript   transaction_simple.ScriptType `json:"txScript"`
			Sender     string                        `json:"sender"`
			Nonce      uint64                        `json:"nonce"`
			Extra      wizard.WizardTxSimpleExtra    `json:"extra"`
			Data       *wizard.WizardTransactionData `json:"data"`
			Fee        *wizard.WizardTransactionFee  `json:"fee"`
			FeeVersion bool                          `json:"feeVersion"`
			Height     uint64                        `json:"height"`
		}{}

		//read txScript
		txScript := &struct {
			TxScript transaction_simple.ScriptType `json:"txScript"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], txScript); err != nil {
			return nil, err
		}

		switch txData.TxScript {
		case transaction_simple.SCRIPT_UPDATE_ASSET_FEE_LIQUIDITY:
			txData.Extra = &wizard.WizardTxSimpleExtraUpdateAssetFeeLiquidity{}
		case transaction_simple.SCRIPT_RESOLUTION_PAY_IN_FUTURE:
			txData.Extra = &wizard.WizardTxSimpleExtraResolutionPayInFuture{}
		default:
			txData.Extra = nil
			return nil, errors.New("Invalid Tx Simple Script")
		}

		if err := webassembly_utils.UnmarshalBytes(args[0], txData); err != nil {
			return nil, err
		}

		transfer := &wizard.WizardTxSimpleTransfer{
			txData.Extra,
			txData.Data,
			txData.Fee,
			txData.Nonce,
			nil,
		}

		if txData.Sender != "" {
			senderWalletAddr, err := app.Wallet.GetWalletAddressByEncodedAddress(txData.Sender, true)
			if err != nil {
				return nil, err
			}

			if senderWalletAddr.PrivateKey.Key == nil {
				return nil, errors.New("Can't be used for transactions as the private key is missing")
			}
			transfer.Key = senderWalletAddr.PrivateKey.Key
		}

		tx, err := wizard.CreateSimpleTx(transfer, true, func(status string) {
			args[1].Invoke(status)
		})
		if err != nil {
			return nil, err
		}

		txJson, err := json.Marshal(tx)
		if err != nil {
			return nil, err
		}

		return []any{
			webassembly_utils.ConvertBytes(txJson),
			webassembly_utils.ConvertBytes(tx.Bloom.Serialized),
		}, nil

	})
}

func signResolutionPayInFuture(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		data := &struct {
			TxId         []byte `json:"txId"`
			PayloadIndex byte   `json:"payloadIndex"`
			Resolution   bool   `json:"resolution"`
			PrivateKey   []byte `json:"privateKey"`
		}{}

		if err := webassembly_utils.UnmarshalBytes(args[0], data); err != nil {
			return nil, err
		}

		key, err := addresses.NewPrivateKey(data.PrivateKey)
		if err != nil {
			return nil, err
		}

		extra := &transaction_simple_extra.TransactionSimpleExtraResolutionPayInFuture{
			nil,
			data.TxId,
			data.PayloadIndex,
			data.Resolution,
			nil, nil,
		}

		signature, err := crypto.SignMessage(extra.MessageForSigning(), data.PrivateKey)
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(struct {
			PublicKey []byte `json:"publicKey"`
			Signature []byte `json:"signature"`
		}{key.GeneratePublicKey(), signature})

	})
}
