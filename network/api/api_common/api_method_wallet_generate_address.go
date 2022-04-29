package api_common

import (
	"errors"
	"net/http"
	"pandora-pay/addresses"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
)

type APIWalletGenerateAddressRequest struct {
	api_types.APIAccountBaseRequest
	PaymentID     helpers.Base64 `json:"paymentID" msgpack:"paymentID"`
	PaymentAmount uint64         `json:"paymentAmount" msgpack:"paymentAmount"`
	PaymentAsset  helpers.Base64 `json:"paymentAsset" msgpack:"paymentAsset"`
}

type APIWalletGenerateAddressReply struct {
	Address string `json:"address" msgpack:"address"`
}

func (api *APICommon) GetWalletGenerateAddress(r *http.Request, args *APIWalletGenerateAddressRequest, reply *APIWalletGenerateAddressReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKey, err := args.GetPublicKeyHash(true)
	if err != nil {
		return err
	}

	walletAddr := api.wallet.GetWalletAddressByPublicKeyHash(publicKey, true)
	if walletAddr == nil {
		return errors.New("address doesn't exist in your waallet")
	}

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		var addr *addresses.Address

		if addr, err = walletAddr.PrivateKey.GenerateAddress(args.PaymentID, args.PaymentAmount, args.PaymentAsset); err != nil {
			return
		}

		reply.Address = addr.EncodeAddr()

		return nil
	})

}
