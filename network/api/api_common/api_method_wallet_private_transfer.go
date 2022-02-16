package api_common

import (
	"context"
	"errors"
	"net/http"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/txs_builder"
)

type APIWalletPrivateTransferRequest struct {
	Data      *txs_builder.TxBuilderCreateZetherTxData `json:"data" msgpack:"data"`
	Propagate bool                                     `json:"propagate" msgpack:"propagate"`
}

type APIWalletPrivateTransferReply struct {
	Result bool                     `json:"result" msgpack:"result"`
	Tx     *transaction.Transaction `json:"tx" msgpack:"tx"`
}

func (api *APICommon) WalletPrivateTransfer(r *http.Request, args *APIWalletPrivateTransferRequest, reply *APIWalletPrivateTransferReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	if reply.Tx, err = api.txsBuilder.CreateZetherTx(args.Data, args.Propagate, true, true, false, context.Background(), func(string) {}); err != nil {
		return
	}

	reply.Result = true

	return
}
