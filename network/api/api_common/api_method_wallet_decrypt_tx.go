package api_common

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet"
)

type APIWalletDecryptTx struct {
	api_types.APIAuthenticateBaseRequest
	APIWalletDecryptTxBase
}

type APIWalletDecryptTxBase struct {
	Hash []byte `json:"hash" msgpack:"hash"`
	api_types.APIAccountBaseRequest
}

type APIWalletDecryptTxReply struct {
	Decrypted *wallet.DecryptedTx `json:"decrypted" msgpack:"decrypted"`
}

func (api *APICommon) WalletDecryptTx(r *http.Request, args *APIWalletDecryptTxBase, reply *APIWalletDecryptTxReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKey, err := args.GetPublicKey(false)
	if err != nil {
		return
	}

	var txSerialized []byte
	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		txSerialized = reader.Get("tx:" + string(args.Hash))

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

	reply.Decrypted, err = api.wallet.DecryptTx(tx, publicKey)

	return
}

func (api *APICommon) WalletDecryptTx_http(values url.Values) (interface{}, error) {
	args := &APIWalletDecryptTx{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIWalletDecryptTxReply{}
	return reply, api.WalletDecryptTx(nil, &args.APIWalletDecryptTxBase, reply, args.CheckAuthenticated())
}

func (api *APICommon) WalletDecryptTx_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIWalletDecryptTxBase{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIWalletDecryptTxReply{}
	return reply, api.WalletDecryptTx(nil, args, reply, conn.Authenticated.IsSet())
}
