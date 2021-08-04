package webassembly

import (
	"encoding/json"
	"pandora-pay/app"
	"pandora-pay/helpers"
	"syscall/js"
)

func createSimpleTx_Float(this js.Value, args []js.Value) interface{} {
	return promiseFunction(func() (interface{}, error) {

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
		}

		txData := &SimpleTxFloatData{}
		if err := json.Unmarshal([]byte(args[0].String()), txData); err != nil {
			return nil, err
		}

		tx, err := app.TransactionsBuilder.CreateSimpleTx_Float(txData.From, txData.Nonce, txData.Amounts, helpers.ConvertHexBytesArrayToBytesArray(txData.AmountsTokens), txData.Dsts, txData.DstsAmounts, helpers.ConvertHexBytesArrayToBytesArray(txData.DstsTokens), txData.FeeFixed, txData.FeePerByte, txData.FeePerByteAuto, txData.FeeToken, txData.PropagateTx, false, false)
		if err != nil {
			return nil, err
		}

		return convertJSON(tx)
	})
}
