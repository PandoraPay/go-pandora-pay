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
	Address       *api_types.APIAccountBaseRequest `json:"address" msgpack:"address"`
	PaymentID     helpers.Base64                   `json:"paymentID" msgpack:"paymentID"`
	PaymentAmount uint64                           `json:"paymentAmount" msgpack:"paymentAmount"`
	PaymentAsset  helpers.Base64                   `json:"paymentAsset" msgpack:"paymentAsset"`
}

type APIWalletGenerateAddressReply struct {
	Address string `json:"address" msgpack:"address"`
}

func (api *APICommon) GetWalletGenerateAddress(r *http.Request, args *APIWalletGenerateAddressRequest, reply *APIWalletGenerateAddressReply, authenticated bool) (err error) {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	publicKey, err := args.Address.GetPublicKey(true)
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
			addr, err = addresses.CreateAddr(publicKey, walletAddr.Staked, walletAddr.SpendPublicKey, walletAddr.Registration, args.PaymentID, args.PaymentAmount, args.PaymentAsset)
		} else {
			addr, err = addresses.CreateAddr(publicKey, false, nil, nil, args.PaymentID, args.PaymentAmount, args.PaymentAsset)
		}
		if err != nil {
			return
		}

		reply.Address = addr.EncodeAddr()

		return nil
	})

}
