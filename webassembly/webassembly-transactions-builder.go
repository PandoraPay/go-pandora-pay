package webassembly

import (
	"encoding/json"
	"errors"
	"pandora-pay/app"
	"pandora-pay/helpers"
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
			From           []string           `json:"from"`
			Nonce          uint64             `json:"nonce"`
			Amounts        []float64          `json:"amounts"`
			AmountsTokens  []helpers.HexBytes `json:"amountsTokens"`
			Dsts           []string           `json:"dsts"`
			DstsAmounts    []float64          `json:"dstsAmounts"`
			DstsTokens     []helpers.HexBytes `json:"dstsTokens"`
			FeeFixed       float64            `json:"feeFixed"`
			FeePerByte     float64            `json:"feePerByte"`
			FeePerByteAuto bool               `json:"feePerByteAuto"`
			FeeToken       helpers.HexBytes   `json:"feeToken"`
			PropagateTx    bool               `json:"propagateTx"`
			AwaitAnswer    bool               `json:"awaitAnswer"`
		}

		txData := &SimpleTxFloatData{}
		if err := json.Unmarshal([]byte(args[0].String()), txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateSimpleTx_Float(txData.From, txData.Nonce, txData.Amounts, helpers.ConvertHexBytesArrayToBytesArray(txData.AmountsTokens), txData.Dsts, txData.DstsAmounts, helpers.ConvertHexBytesArrayToBytesArray(txData.DstsTokens), txData.FeeFixed, txData.FeePerByte, txData.FeePerByteAuto, txData.FeeToken, txData.PropagateTx, txData.AwaitAnswer, false, func(status string) {
			callback.Invoke(status)
			time.Sleep(10 * time.Millisecond)
		})
		if err != nil {
			return nil, err
		}

		return convertJSON(tx)
	})
}
