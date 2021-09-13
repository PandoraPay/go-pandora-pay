package webassembly

import (
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/config"
	"pandora-pay/helpers"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/transactions-builder/wizard"
	"syscall/js"
	"time"
)

func createZetherTx_Float(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		if len(args) != 3 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeFunction || args[2].Type() != js.TypeString {
			return nil, errors.New("Argument must be a string and a callback")
		}

		if err := app.Wallet.Encryption.CheckPassword(args[2].String(), false); err != nil {
			return nil, err
		}

		type SimpleTxFloatData struct {
			From        []string                                          `json:"from"`
			Nonce       uint64                                            `json:"nonce"`
			Token       helpers.HexBytes                                  `json:"token"`
			Amounts     []float64                                         `json:"amounts"`
			Dsts        []string                                          `json:"dsts"`
			DstsAmounts []float64                                         `json:"dstsAmounts"`
			Data        *wizard.TransactionsWizardData                    `json:"data"`
			Fee         *transactions_builder.TransactionsBuilderFeeFloat `json:"fee"`
			PropagateTx bool                                              `json:"propagateTx"`
			AwaitAnswer bool                                              `json:"awaitAnswer"`
		}

		txData := &SimpleTxFloatData{}
		if err := json.Unmarshal([]byte(args[0].String()), txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.createZetherTx_Float(txData.From, txData.Nonce, config.NATIVE_TOKEN_FULL, txData.Amounts, txData.Dsts, txData.DstsAmounts, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, func(status string) {
			args[1].Invoke(status)
			time.Sleep(10 * time.Millisecond)
		})
		if err != nil {
			return nil, err
		}

		return convertJSON(tx)
	})
}

func createUpdateDelegateTx_Float(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

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
			DelegateNewFee               uint16                                            `json:"delegateNewFee"`
			Data                         *wizard.TransactionsWizardData                    `json:"data"`
			Fee                          *transactions_builder.TransactionsBuilderFeeFloat `json:"fee"`
			PropagateTx                  bool                                              `json:"propagateTx"`
			AwaitAnswer                  bool                                              `json:"awaitAnswer"`
		}

		txData := &DelegateTxFloatData{}
		if err := json.Unmarshal([]byte(args[0].String()), txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateUpdateDelegateTx_Float(txData.From, txData.Nonce, txData.DelegateNewPublicKeyGenerate, txData.DelegateNewPubKey, txData.DelegateNewFee, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, func(status string) {
			args[1].Invoke(status)
			time.Sleep(10 * time.Millisecond)
		})
		if err != nil {
			return nil, err
		}

		return convertJSON(tx)

	})
}

func createUnstakeTx_Float(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

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

		tx, err := app.TransactionsBuilder.CreateUnstakeTx_Float(txData.From, txData.Nonce, txData.UnstakeAmount, txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, func(status string) {
			args[1].Invoke(status)
			time.Sleep(10 * time.Millisecond)
		})
		if err != nil {
			return nil, err
		}

		return convertJSON(tx)

	})
}
