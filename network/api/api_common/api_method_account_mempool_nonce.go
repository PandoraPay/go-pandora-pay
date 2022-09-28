package api_common

import (
	"net/http"
	"pandora-pay/network/api/api_common/api_types"
)

type APIAccountMempoolNonceRequest struct {
	api_types.APIAccountBaseRequest
}

type APIAccountMempoolNonceReply struct {
	Nonce uint64 `json:"nonce" msgpack:"nonce"`
}

func (api *APICommon) GetAccountMempoolNonce(r *http.Request, args *APIAccountMempoolNonceRequest, reply *APIAccountMempoolNonceReply) error {
	publicKey, err := args.GetPublicKey(true)
	if err != nil {
		return err
	}

	//if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) error {
	//
	//	chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
	//	plainAccs := plain_accounts.NewPlainAccounts(reader)
	//
	//	plainAcc, err := plainAccs.Get(publicKey)
	//	if err != nil {
	//		return err
	//	}
	//	if plainAcc != nil {
	//		reply.Nonce = plainAcc.Nonce
	//	}
	//
	//	return nil
	//}); err != nil {
	//	return err
	//}

	reply.Nonce = api.mempool.GetNonce(publicKey, reply.Nonce)
	return nil
}
