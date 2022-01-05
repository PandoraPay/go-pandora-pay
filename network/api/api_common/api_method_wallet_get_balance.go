package api_common

import (
	"encoding/json"
	"errors"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIWalletGetBalance struct {
	api_types.APIAuthenticateBaseRequest
	APIWalletGetBalanceBase
}

type APIWalletGetBalanceBase struct {
	List []*api_types.APIAccountBaseRequest
}

type APIWalletGetBalanceReply struct {
	BalancePlainAcc uint64                          `json:"balancePlainAcc"`
	Balances        []*APIWalletGetBalanceDataReply `json:"balance"`
}

type APIWalletGetBalanceDataReply struct {
	Amount uint64           `json:"amount"`
	Asset  helpers.HexBytes `json:"asset"`
}

func (api *APICommon) WalletGetBalance(r *http.Request, args *APIWalletGetBalanceBase, reply *APIWalletGetBalanceReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKeys := make([][]byte, len(args.List))
	for i, it := range args.List {
		if publicKeys[i], err = it.GetPublicKey(); err != nil {
			return
		}
	}
	//
	//if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
	//	chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
	//	dataStorage := data_storage.NewDataStorage(reader)
	//
	//	for i, publicKey := range publicKeys {
	//		var assetsList [][]byte
	//		if assetsList, err = dataStorage.AccsCollection.GetAccountAssets(publicKey); err != nil {
	//			return
	//		}
	//
	//		var plainAcc *plain_account.PlainAccount
	//		if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(publicKey, chainHeight); err != nil {
	//			return
	//		}
	//	}
	//
	//	if plainAcc != nil {
	//
	//	}
	//
	//}; err != nil {
	//	return err
	//}

	return nil
}

func (api *APICommon) WalletGetBalance_http(values url.Values) (interface{}, error) {
	args := &APIWalletGetBalance{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIWalletGetBalanceReply{}
	return reply, api.WalletGetBalance(nil, &args.APIWalletGetBalanceBase, reply, args.CheckAuthenticated())
}

func (api *APICommon) WalletGetBalance_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIWalletGetBalanceBase{}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIWalletGetBalanceReply{}
	return reply, api.WalletGetBalance(nil, args, reply, conn.Authenticated.IsSet())
}
