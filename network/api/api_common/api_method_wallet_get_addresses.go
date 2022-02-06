package api_common

import (
	"errors"
	"net/http"
	"net/url"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/wallet"
	"pandora-pay/wallet/wallet_address"
)

type APIWalletGetAccounts struct {
	api_types.APIAuthenticateBaseRequest
}

type APIWalletGetAccountsReply struct {
	Version   wallet.Version                  `json:"version" msgpack:"version"`
	Encrypted wallet.EncryptedVersion         `json:"encrypted" msgpack:"encrypted"`
	Addresses []*wallet_address.WalletAddress `json:"addresses" msgpack:"addresses"`
}

func (api *APICommon) WalletGetAddresses(r *http.Request, args *struct{}, reply *APIWalletGetAccountsReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	api.wallet.Lock.RLock()
	defer api.wallet.Lock.RUnlock()

	reply.Version = api.wallet.Version
	reply.Encrypted = api.wallet.Encryption.Encrypted

	reply.Addresses = make([]*wallet_address.WalletAddress, len(api.wallet.Addresses))
	for i, addr := range api.wallet.Addresses {
		if reply.Addresses[i], err = generics.Clone[*wallet_address.WalletAddress](addr, new(wallet_address.WalletAddress)); err != nil {
			return
		}
	}

	return
}

func (api *APICommon) WalletGetAddresses_http(values url.Values) (interface{}, error) {
	args := &APIWalletGetAccounts{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIWalletGetAccountsReply{}
	return reply, api.WalletGetAddresses(nil, &struct{}{}, reply, args.CheckAuthenticated())
}

func (api *APICommon) WalletGetAddresses_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &APIWalletGetAccountsReply{}
	return reply, api.WalletGetAddresses(nil, &struct{}{}, reply, conn.Authenticated.IsSet())
}
