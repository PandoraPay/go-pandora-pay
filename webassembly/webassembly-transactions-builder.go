package webassembly

import (
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/helpers"
	transactions_builder "pandora-pay/transactions-builder"
	"pandora-pay/transactions-builder/wizard"
	"syscall/js"
	"time"
)

func createSimpleTx_Float(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

		if len(args) != 2 || args[0].Type() != js.TypeString || args[1].Type() != js.TypeFunction {
			return nil, errors.New("Argument must be a string and a callback")
		}

		callback := args[1]

		type SimpleTxFloatData struct {
			From          []string                                          `json:"from"`
			Nonce         uint64                                            `json:"nonce"`
			Amounts       []float64                                         `json:"amounts"`
			AmountsTokens []helpers.HexBytes                                `json:"amountsTokens"`
			Dsts          []string                                          `json:"dsts"`
			DstsAmounts   []float64                                         `json:"dstsAmounts"`
			DstsTokens    []helpers.HexBytes                                `json:"dstsTokens"`
			Data          *wizard.TransactionsWizardData                    `json:"data"`
			Fee           *transactions_builder.TransactionsBuilderFeeFloat `json:"fee"`
			PropagateTx   bool                                              `json:"propagateTx"`
			AwaitAnswer   bool                                              `json:"awaitAnswer"`
		}

		txData := &SimpleTxFloatData{}
		if err := json.Unmarshal([]byte(args[0].String()), txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateSimpleTx_Float(txData.From, txData.Nonce, txData.Amounts, helpers.ConvertHexBytesArrayToBytesArray(txData.AmountsTokens), txData.Dsts, txData.DstsAmounts, helpers.ConvertHexBytesArrayToBytesArray(txData.DstsTokens), txData.Data, txData.Fee, txData.PropagateTx, txData.AwaitAnswer, false, func(status string) {
			callback.Invoke(status)
			time.Sleep(10 * time.Millisecond)
		})
		if err != nil {
			return nil, err
		}

		return convertJSON(tx)
	})
}
