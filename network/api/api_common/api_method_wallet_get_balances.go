package api_common

import (
	"encoding/binary"
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
	Balances []*APIWalletGetBalanceDataReply `json:"balance" msgpack:"balance"`
}

type APIWalletGetBalanceDataReply struct {
	Balance []byte `json:"balance" msgpack:"balance"`
	Amount  uint64 `json:"amount" msgpack:"amount"`
	Asset   []byte `json:"asset" msgpack:"asset"`
}

func (api *APICommon) GetWalletBalances(r *http.Request, args *APIWalletGetBalanceRequest, reply *APIWalletGetBalancesReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKeys := make([][]byte, len(args.List))
	for i, it := range args.List {
		if publicKeys[i], err = it.GetPublicKey(true); err != nil {
			return
		}
	}

	walletAddresses := make([]*wallet_address.WalletAddress, len(publicKeys))
	for i, publicKey := range publicKeys {
		if walletAddresses[i] = api.wallet.GetWalletAddressByPublicKey(publicKey, false); walletAddresses[i] == nil {
			return errors.New(fmt.Sprintf("input %d doesn't exist in your wallet", i))
		}
	}

	reply.Results = make([]*APIWalletGetBalancesResultReply, len(publicKeys))

	if err = store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {
		chainHeight, _ := binary.Uvarint(reader.Get("chainHeight"))
		dataStorage := data_storage.NewDataStorage(reader)

		for i, publicKey := range publicKeys {

			var isReg bool
			if isReg, err = dataStorage.Regs.Exists(string(publicKey)); err != nil {
				return
			}

			reply.Results[i] = &APIWalletGetBalancesResultReply{}

			reply.Results[i].Address = walletAddresses[i].GetAddress(isReg)

			var plainAcc *plain_account.PlainAccount
			if plainAcc, err = dataStorage.PlainAccs.GetPlainAccount(publicKey, chainHeight); err != nil {
				return
			}

			var assetsList [][]byte
			if assetsList, err = dataStorage.AccsCollection.GetAccountAssets(publicKey); err != nil {
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
				if acc, err = accs.GetAccount(walletAddresses[i].PublicKey, chainHeight); err != nil {
					return
				}

				reply.Results[i].Balances[j] = &APIWalletGetBalanceDataReply{
					acc.Balance.Amount.Serialize(),
					0,
					assetId,
				}

			}

		}

		return
	}); err != nil {
		return
	}

	for i, publicKey := range publicKeys {
		for _, data := range reply.Results[i].Balances {

			if data.Amount, err = api.wallet.DecryptBalanceByPublicKey(publicKey, data.Balance, data.Asset, false, 0, true, true, nil, func(status string) {}); err != nil {
				return
			}
		}
	}

	return
}
