package webassembly

import (
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/helpers"
	"pandora-pay/transactions_builder"
	"pandora-pay/transactions_builder/wizard"
	"pandora-pay/webassembly/webassembly_utils"
	"syscall/js"
)

func createUpdateDelegateTx_Float(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		type DelegateTxFloatData struct {
			From                         string                                            `json:"from"`
			Nonce                        uint64                                            `json:"nonce"`
			DelegateNewPublicKeyGenerate bool                                              `json:"delegateNewPublicKeyGenerate"`
			DelegateNewPubKey            helpers.HexBytes                                  `json:"delegateNewPubKey"`
			DelegateNewFee               uint64                                            `json:"delegateNewFee"`
			Data                         *wizard.TransactionsWizardData                    `json:"data"`
			Fee                          *transactions_builder.TransactionsBuilderFeeFloat `json:"fee"`
			PropagateTx                  bool                                              `json:"propagateTx"`
			AwaitAnswer                  bool                                              `json:"awaitAnswer"`
		}

		txData := &DelegateTxFloatData{}
		if err := json.Unmarshal([]byte(args[0].String()), txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateUpdateDelegateTx_Float(txData.From, txData.Nonce, txData.DelegateNewPublicKeyGenerate, txData.DelegateNewPubKey, txData.DelegateNewFee, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, false, func(status string) {
			args[1].Invoke(status)
		})
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(tx)

	})
}

func createUnstakeTx_Float(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		type DelegateTxFloatData struct {
			From          string                                            `json:"from"`
			Nonce         uint64                                            `json:"nonce"`
			UnstakeAmount float64                                           `json:"unstakeAmount"`
			Data          *wizard.TransactionsWizardData                    `json:"data"`
			Fee           *transactions_builder.TransactionsBuilderFeeFloat `json:"fee"`
			PropagateTx   bool                                              `json:"propagateTx"`
			AwaitAnswer   bool                                              `json:"awaitAnswer"`
		}

		txData := &DelegateTxFloatData{}
		if err := json.Unmarshal([]byte(args[0].String()), txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateUnstakeTx_Float(txData.From, txData.Nonce, txData.UnstakeAmount, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, false, func(status string) {
			args[1].Invoke(status)
		})
		if err != nil {
			return nil, err
		}

		return webassembly_utils.ConvertJSONBytes(tx)

	})
}
