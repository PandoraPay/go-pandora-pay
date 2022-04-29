package api_common

import (
	"net/http"
	"pandora-pay/blockchain/data_storage/plain_accounts"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIAccountMempoolNonceRequest struct {
	api_types.APIAccountBaseRequest
}

type APIAccountMempoolNonceReply struct {
	Nonce uint64 `json:"nonce" msgpack:"nonce"`
}

func (api *APICommon) GetAccountMempoolNonce(r *http.Request, args *APIAccountMempoolNonceRequest, reply *APIAccountMempoolNonceReply) error {
	publicKeyHash, err := args.GetPublicKeyHash(true)
	if err != nil {
		return err
	}

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) error {

		plainAccs := plain_accounts.NewPlainAccounts(reader)

		plainAcc, err := plainAccs.GetPlainAccount(publicKeyHash)
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

	reply.Nonce = api.mempool.GetNonce(publicKeyHash, reply.Nonce)
	return nil
}
