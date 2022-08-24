package routes

import (
	"context"
	"encoding/json"
	"pandora-pay/builds/builds_data"
	"pandora-pay/txs_builder/wizard"
)

func RouteTransactionsBuilderCreateZetherTx(req *builds_data.TransactionsBuilderCreateZetherTxReq) (any, error) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, publicKeyIndexes, feesFinal, err := builds_data.PrepareData(req)
	if err != nil {
		return nil, err
	}

	tx, err := wizard.CreateZetherTx(transfers, emap, hasRollovers, ringsSenderMembers, ringsRecipientMembers, req.ChainKernelHeight, req.ChainKernelHash, publicKeyIndexes, feesFinal, ctx, func(status string) {})
	if err != nil {
		return nil, err
	}

	txJson, err := json.Marshal(tx)
	if err != nil {
		return nil, err
	}

	return []interface{}{
		txJson,
		tx.Bloom.Serialized,
	}, nil
}
