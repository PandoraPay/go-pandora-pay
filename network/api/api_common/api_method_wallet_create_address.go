package api_common

import (
	"errors"
	"net/http"
	"pandora-pay/helpers/generics"
	"pandora-pay/wallet/wallet_address"
)

type APIWalletCreateAddressRequest struct {
	Name          string `json:"name" msgpack:"name"`
}

type APIWalletCreateAddressReply struct {
	Address *wallet_address.WalletAddress `json:"address" msgpack:"address"`
}

func (api *APICommon) GetWalletCreateAddress(r *http.Request, args *APIWalletCreateAddressRequest, reply *APIWalletCreateAddressReply, authenticated bool) error {
	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	addr, err := api.wallet.AddNewAddress(true, args.Name, true)
	if err != nil {
		return err
	}

	reply.Address, err = generics.Clone[*wallet_address.WalletAddress](addr, new(wallet_address.WalletAddress))
	return nil
}
