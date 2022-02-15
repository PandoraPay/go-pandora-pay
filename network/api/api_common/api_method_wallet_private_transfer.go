package api_common

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"io"
	"net/http"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/txs_builder"
)

type APIWalletPrivateTransfer struct {
	api_types.APIAuthenticateBaseRequest
	APIWalletPrivateTransferBase
}

type APIWalletPrivateTransferBase struct {
	Data      *txs_builder.TxBuilderCreateZetherTxData `json:"data" msgpack:"data"`
	Propagate bool                                     `json:"propagate" msgpack:"propagate"`
}

type APIWalletPrivateTransferReply struct {
	Result bool                     `json:"result" msgpack:"result"`
	Tx     *transaction.Transaction `json:"tx" msgpack:"tx"`
}

func (api *APICommon) WalletPrivateTransfer(r *http.Request, args *APIWalletPrivateTransferBase, reply *APIWalletPrivateTransferReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	if reply.Tx, err = api.txsBuilder.CreateZetherTx(args.Data, args.Propagate, true, true, false, context.Background(), func(string) {}); err != nil {
		return
	}

	reply.Result = true

	return
}

func (api *APICommon) WalletPrivateTransfer_http(values io.ReadCloser) (interface{}, error) {
	args := &APIWalletPrivateTransfer{}
	if err := json.NewDecoder(values).Decode(args); err != nil {
		return nil, err
	}
	reply := &APIWalletPrivateTransferReply{}
	return reply, api.WalletPrivateTransfer(nil, &args.APIWalletPrivateTransferBase, reply, args.CheckAuthenticated())
}

func (api *APICommon) WalletPrivateTransfer_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIWalletPrivateTransferBase{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIWalletPrivateTransferReply{}
	return reply, api.WalletPrivateTransfer(nil, args, reply, conn.Authenticated.IsSet())
}
