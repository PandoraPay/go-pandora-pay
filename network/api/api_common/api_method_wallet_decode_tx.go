package api_common

import (
	"context"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIWalletDecodeTx struct {
	api_types.APIAuthenticateBaseRequest
	APIWalletDecodeTxBase
}

type APIWalletDecodeTxBase struct {
	TxHash helpers.HexBytes `json:"txHash" msgpack:"txHash"`
}

type APIWalletDecodeTxReply struct {
	Type transaction_type.TransactionVersion `json:"type" msgpack:"type"`
}

func (api *APICommon) WalletDecodeTx(r *http.Request, args *APIWalletDecodeTxBase, reply *APIWalletDecodeTxReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	var txSerialized []byte
	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		txSerialized = reader.Get("tx:" + string(args.TxHash))

		return
	}); err != nil {
		return
	}

	if len(txSerialized) == 0 {
		return errors.New("Tx was not found in the storage")
	}

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(txSerialized)); err != nil {
		return
	}

	reply.Type = tx.Version

	switch tx.Version {
	case transaction_type.TX_ZETHER:
		err = api.wallet.DecodeZetherTx(tx, context.Background())
	}

	return
}

func (api *APICommon) WalletDecodeTx_http(values url.Values) (interface{}, error) {
	args := &APIWalletDecodeTx{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIWalletDecodeTxReply{}
	return reply, api.WalletDecodeTx(nil, &args.APIWalletDecodeTxBase, reply, args.CheckAuthenticated())
}

func (api *APICommon) WalletDecodeTx_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIWalletDecodeTxBase{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIWalletDecodeTxReply{}
	return reply, api.WalletDecodeTx(nil, args, reply, conn.Authenticated.IsSet())
}
