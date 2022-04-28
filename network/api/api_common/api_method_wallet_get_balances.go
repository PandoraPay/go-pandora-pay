package api_common

import (
	"errors"
	"fmt"
	"net/http"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
)

type APIWalletGetBalanceRequest struct {
	List []*api_types.APIAccountBaseRequest `json:"list" msgpack:"list"`
}

type APIWalletGetBalancesReply struct {
	Results []*APIWalletGetBalancesResultReply `json:"results" msgpack:"results"`
}

type APIWalletGetBalancesResultReply struct {
	Address  string                          `json:"address" msgpack:"address"`
	PlainAcc *plain_account.PlainAccount     `json:"plainAcc" msgpack:"plainAcc"`
	Balances []*APIWalletGetBalanceDataReply `json:"balances" msgpack:"balances"`
}

type APIWalletGetBalanceDataReply struct {
	Balance uint64 `json:"balance" msgpack:"balance"`
	Asset   []byte `json:"asset" msgpack:"asset"`
}

func (api *APICommon) GetWalletBalances(r *http.Request, args *APIWalletGetBalanceRequest, reply *APIWalletGetBalancesReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKeyHashes := make([][]byte, len(args.List))
	for i, it := range args.List {
		if publicKeyHashes[i], err = it.GetPublicKeyHash(true); err != nil {
			return
		}
	}

	walletAddresses := make([]*wallet_address.WalletAddress, len(publicKeyHashes))
	for i, publicKey := range publicKeyHashes {
		if walletAddresses[i] = api.wallet.GetWalletAddressByPublicKeyHash(publicKey, true); walletAddresses[i] == nil {
			return errors.New(fmt.Sprintf("input %d doesn't exist in your wallet", i))
		}
	}

	reply.Results = make([]*APIWalletGetBalancesResultReply, len(publicKeyHashes))

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		for i, publicKeyHash := range publicKeyHashes {

			reply.Results[i] = &APIWalletGetBalancesResultReply{}

			reply.Results[i].Address = walletAddresses[i].GetAddress()

			var plainAcc *plain_account.PlainAccount
			if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(publicKeyHash); err != nil {
				return
			}

			var assetsList [][]byte
			if assetsList, err = dataStorage.AccsCollection.GetAccountAssets(publicKeyHash); err != nil {
				return
			}

			reply.Results[i].PlainAcc = plainAcc
			reply.Results[i].Balances = make([]*APIWalletGetBalanceDataReply, len(assetsList))

			for j, assetId := range assetsList {

				var accs *accounts.Accounts
				var acc *account.Account
				if accs, err = dataStorage.AccsCollection.GetMap(assetId); err != nil {
					return
				}
				if acc, err = accs.GetAccount(walletAddresses[i].PublicKey); err != nil {
					return
				}

				reply.Results[i].Balances[j] = &APIWalletGetBalanceDataReply{
					acc.Balance,
					assetId,
				}

			}

		}

		return
	}); err != nil {
		return
	}

	return
}
