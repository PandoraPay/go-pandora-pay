package api_common

import (
	"errors"
	"net/http"
	"pandora-pay/network/api/api_common/api_types"
)

type APIWalletDeleteAddressRequest struct {
	api_types.APIAccountBaseRequest
}

type APIWalletDeleteAddressReply struct {
	Status bool `json:"status" msgpack:"status"`
}

func (api *APICommon) GetWalletDeleteAddress(r *http.Request, args *APIWalletDeleteAddressRequest, reply *APIWalletDeleteAddressReply, authenticated bool) error {
	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKey, err := args.GetPublicKey(true)
	if err != nil {
		return err
	}

	reply.Status, err = api.wallet.RemoveAddressByPublicKey(publicKey, true)
	return err
}
