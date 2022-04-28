package api_common

import (
	"errors"
	"net/http"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage"
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

	publicKey, err := args.GetPublicKey(true)
	if err != nil {
		return err
	}

	walletAddr := api.wallet.GetWalletAddressByPublicKey(publicKey, true)
	if walletAddr == nil {
		return errors.New("address doesn't exist in your waallet")
	}

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		dataStorage := data_storage.NewDataStorage(reader)

		var isReg bool
		if isReg, err = dataStorage.Regs.Exists(string(publicKey)); err != nil {
			return
		}

		var addr *addresses.Address

		if !isReg {
			addr, err = walletAddr.PrivateKey.GenerateAddress(walletAddr.Staked, walletAddr.SpendPublicKey, true, args.PaymentID, args.PaymentAmount, args.PaymentAsset)
		} else {
			addr, err = walletAddr.PrivateKey.GenerateAddress(false, nil, false, args.PaymentID, args.PaymentAmount, args.PaymentAsset)
		}
		if err != nil {
			return
		}

		reply.Address = addr.EncodeAddr()

		return nil
	})

}
