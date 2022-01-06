package api_common

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/accounts"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/cryptography/crypto"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"pandora-pay/wallet/wallet_address"
)

type APIWalletGetBalance struct {
	api_types.APIAuthenticateBaseRequest
	APIWalletGetBalancesBase
}

type APIWalletGetBalancesBase struct {
	List []*api_types.APIAccountBaseRequest `json:"list"  schema:"list"`
}

type APIWalletGetBalancesReply struct {
	Results []*APIWalletGetBalancesResultReply `json:"results"`
}

type APIWalletGetBalancesResultReply struct {
	Address  string                          `json:"address"`
	PlainAcc *plain_account.PlainAccount     `json:"plainAcc"`
	Balances []*APIWalletGetBalanceDataReply `json:"balance"`
}

type APIWalletGetBalanceDataReply struct {
	Balance helpers.HexBytes `json:"amount"`
	Value   uint64           `json:"value"`
	Asset   helpers.HexBytes `json:"asset"`
}

func (api *APICommon) WalletGetBalances(r *http.Request, args *APIWalletGetBalancesBase, reply *APIWalletGetBalancesReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKeys := make([][]byte, len(args.List))
	for i, it := range args.List {
		if publicKeys[i], err = it.GetPublicKey(); err != nil {
			return
		}
	}

	walletAddresses := make([]*wallet_address.WalletAddress, len(publicKeys))
	for i, publicKey := range publicKeys {
		if walletAddresses[i] = api.wallet.GetWalletAddressByPublicKey(publicKey); walletAddresses[i] == nil {
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
				if acc, err = accs.GetAccount(walletAddresses[i].PublicKey); err != nil {
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

			var balancePoint *crypto.ElGamal
			if balancePoint, err = new(crypto.ElGamal).Deserialize(data.Balance); err != nil {
				return
			}

			if data.Value, err = api.wallet.DecodeBalanceByPublicKey(publicKey, balancePoint, data.Asset, true, true, nil, func(status string) {}); err != nil {
				return
			}
		}
	}

	return
}

func (api *APICommon) WalletGetBalances_http(values url.Values) (interface{}, error) {
	args := &APIWalletGetBalance{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIWalletGetBalancesReply{}
	return reply, api.WalletGetBalances(nil, &args.APIWalletGetBalancesBase, reply, args.CheckAuthenticated())
}

func (api *APICommon) WalletGetBalances_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIWalletGetBalancesBase{}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIWalletGetBalancesReply{}
	return reply, api.WalletGetBalances(nil, args, reply, conn.Authenticated.IsSet())
}
