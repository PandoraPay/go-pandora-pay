package api_common

import (
	"encoding/binary"
	"encoding/json"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountMempoolNonceRequest struct {
	api_types.APIAccountBaseRequest
}

func (api *APICommon) AccountMempoolNonce(r *http.Request, args *APIAccountMempoolNonceRequest, reply *uint64) error {
	publicKey, err := args.GetPublicKey()
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
			*reply = plainAcc.Nonce
		}

		return nil
	}); err != nil {
		return err
	}

	*reply = api.mempool.GetNonce(publicKey, *reply)
	return nil
}

func (api *APICommon) GetAccountMempoolNonce_http(values url.Values) (interface{}, error) {
	args := &APIAccountMempoolNonceRequest{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	var reply uint64
	return reply, api.AccountMempoolNonce(nil, args, &reply)
}

func (api *APICommon) GetAccountMempoolNonce_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIAccountMempoolNonceRequest{}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	var reply uint64
	return reply, api.AccountMempoolNonce(nil, args, &reply)
}
