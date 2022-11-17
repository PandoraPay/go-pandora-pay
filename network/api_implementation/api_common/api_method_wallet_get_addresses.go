package api_common

import (
	"errors"
	"net/http"
	"pandora-pay/helpers/generics"
	"pandora-pay/wallet"
	"pandora-pay/wallet/wallet_address"
)

type APIWalletGetAccountsReply struct {
	Version   wallet.Version                  `json:"version" msgpack:"version"`
	Encrypted wallet.EncryptedVersion         `json:"encrypted" msgpack:"encrypted"`
	Addresses []*wallet_address.WalletAddress `json:"addresses" msgpack:"addresses"`
}

func (api *APICommon) GetWalletAddresses(r *http.Request, args *struct{}, reply *APIWalletGetAccountsReply, authenticated bool) (err error) {

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
