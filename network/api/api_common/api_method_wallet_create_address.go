package api_common

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/wallet/wallet_address"
)

type APIWalletCreateAddress struct {
	api_types.APIAuthenticateBaseRequest
	APIWalletCreateAddressBase
}

type APIWalletCreateAddressBase struct {
	Name string `json:"name" msgpack:"name"`
}

type APIWalletCreateAddressReply struct {
	Address *wallet_address.WalletAddress `json:"address" msgpack:"address"`
}

func (api *APICommon) WalletCreateAddress(r *http.Request, args *APIWalletCreateAddressBase, reply *APIWalletCreateAddressReply, authenticated bool) error {
	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	addr, err := api.wallet.AddNewAddress(true, args.Name)
	if err != nil {
		return err
	}

	reply.Address, err = generics.Clone[*wallet_address.WalletAddress](addr, new(wallet_address.WalletAddress))
	return nil
}

func (api *APICommon) WalletCreateAddress_http(values url.Values) (interface{}, error) {
	args := &APIWalletCreateAddress{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIWalletCreateAddressReply{}
	return reply, api.WalletCreateAddress(nil, &args.APIWalletCreateAddressBase, reply, args.CheckAuthenticated())
}

func (api *APICommon) WalletCreateAddress_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	args := &APIWalletCreateAddressBase{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIWalletCreateAddressReply{}
	return reply, api.WalletCreateAddress(nil, args, reply, conn.Authenticated.IsSet())
}
