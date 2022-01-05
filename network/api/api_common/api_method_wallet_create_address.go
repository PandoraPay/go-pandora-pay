package api_common

import (
	"errors"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIWalletCreateAddress struct {
	api_types.APIAuthenticateBaseRequest
}

type APIWalletCreateAddressReply struct {
	Address *APIWalletReplyAddress
}

func (api *APICommon) WalletCreateAddress(r *http.Request, args *struct{}, reply *APIWalletCreateAddressReply, authenticated bool) error {
	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	addr, err := api.wallet.AddNewAddress(true)
	if err != nil {
		return err
	}

	reply.Address = importWalletAddress(addr)
	return nil
}

func (api *APICommon) WalletCreateAddress_http(values url.Values) (interface{}, error) {
	args := &APIWalletGetAccounts{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIWalletCreateAddressReply{}
	return reply, api.WalletCreateAddress(nil, &struct{}{}, reply, args.CheckAuthenticated())
}

func (api *APICommon) WalletCreateAddress_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &APIWalletCreateAddressReply{}
	return reply, api.WalletCreateAddress(nil, &struct{}{}, reply, conn.Authenticated.IsSet())
}
