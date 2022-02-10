package api_common

import (
	"encoding/binary"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountMempoolNonceRequest struct {
	api_types.APIAccountBaseRequest
}

type APIAccountMempoolNonceReply struct {
	Nonce uint64 `json:"nonce" msgpack:"nonce"`
}

func (api *APICommon) AccountMempoolNonce(r *http.Request, args *APIAccountMempoolNonceRequest, reply *APIAccountMempoolNonceReply) error {
	publicKey, err := args.GetPublicKey(true)
	if err != nil {
		return err
	}

	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) error {

		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
		plainAccs := plain_accounts.NewPlainAccounts(reader)

		plainAcc, err := plainAccs.GetPlainAccount(publicKey, chainHeight)
		if err != nil {
			return err
		}
		if plainAcc != nil {
			reply.Nonce = plainAcc.Nonce
		}

		return nil
	}); err != nil {
		return err
	}

	reply.Nonce = api.mempool.GetNonce(publicKey, reply.Nonce)
	return nil
}

func (api *APICommon) GetAccountMempoolNonce_http(values url.Values) (interface{}, error) {
	args := &APIAccountMempoolNonceRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIAccountMempoolNonceReply{}
	return reply, api.AccountMempoolNonce(nil, args, reply)
}

func (api *APICommon) GetAccountMempoolNonce_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAccountMempoolNonceRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIAccountMempoolNonceReply{}
	return reply, api.AccountMempoolNonce(nil, args, reply)
}
