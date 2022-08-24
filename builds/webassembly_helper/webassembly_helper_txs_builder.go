package main

import (
	"context"
	"encoding/json"
	"errors"
	"pandora-pay/builds/builds_data"
	"pandora-pay/builds/webassembly/webassembly_utils"
	"pandora-pay/txs_builder/wizard"
	"syscall/js"
)

func createZetherTx(this js.Value, args []js.Value) interface{} {
	return webassembly_utils.PromiseFunction(func() (interface{}, error) {

		if len(args) != 2 || args[0].Type() != js.TypeObject || args[1].Type() != js.TypeFunction {
			return nil, errors.New("Argument must be a string and a callback")
		}

		txData := &builds_data.TransactionsBuilderCreateZetherTxReq{}
		if err := webassembly_utils.UnmarshalBytes(args[0], txData); err != nil {
			return nil, err
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, publicKeyIndexes, feesFinal, err := builds_data.PrepareData(txData)
		if err != nil {
			return nil, err
		}

		tx, err := wizard.CreateZetherTx(transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, txData.ChainKernelHeight, txData.ChainKernelHash, publicKeyIndexes, feesFinal, ctx, func(status string) {
			args[1].Invoke(status)
		})
		if err != nil {
			return nil, err
		}

		txJson, err := json.Marshal(tx)
		if err != nil {
			return nil, err
		}

		return []interface{}{
			webassembly_utils.ConvertBytes(txJson),
			webassembly_utils.ConvertBytes(tx.Bloom.Serialized),
		}, nil
	})
}
