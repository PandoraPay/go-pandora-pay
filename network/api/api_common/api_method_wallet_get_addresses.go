package api_common

import (
	"errors"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/wallet"
	"pandora-pay/wallet/wallet_address"
)

type APIWalletGetAccounts struct {
	api_types.APIAuthenticateBaseRequest
}

type APIWalletGetAccountsReply struct {
	Version   wallet.Version          `json:"version"`
	Encrypted wallet.EncryptedVersion `json:"encrypted"`
	Addresses []*APIWalletReplyAddress
}

type APIWalletReplyAddress struct {
	Version                    wallet_address.Version                          `json:"version"`
	Name                       string                                          `json:"name"`
	SeedIndex                  uint32                                          `json:"seedIndex"`
	IsMine                     bool                                            `json:"isMine"`
	PrivateKey                 helpers.HexBytes                                `json:"privateKey"`
	Registration               helpers.HexBytes                                `json:"registration"`
	PublicKey                  helpers.HexBytes                                `json:"publicKey"`
	AddressEncoded             string                                          `json:"addressEncoded"`
	AddressRegistrationEncoded string                                          `json:"addressRegistrationEncoded"`
	DelegatedStake             *APIWalletGetAccountsReplyAddressDelegatedStake `json:"delegatedStake"`
}

type APIWalletGetAccountsReplyAddressDelegatedStake struct {
	PrivateKey     helpers.HexBytes `json:"privateKey"`
	PublicKey      helpers.HexBytes `json:"publicKey"`
	LastKnownNonce uint32           `json:"lastKnownNonce"`
}

func importWalletAddress(addr *wallet_address.WalletAddress) *APIWalletReplyAddress {
	return &APIWalletReplyAddress{
		addr.Version,
		addr.Name,
		addr.SeedIndex,
		addr.IsMine,
		addr.PrivateKey.Key,
		addr.Registration,
		addr.PublicKey,
		addr.AddressEncoded,
		addr.AddressRegistrationEncoded,
		&APIWalletGetAccountsReplyAddressDelegatedStake{
			addr.DelegatedStake.PrivateKey.Key,
			addr.DelegatedStake.PublicKey,
			addr.DelegatedStake.LastKnownNonce,
		},
	}
}

func (api *APICommon) WalletGetAddresses(r *http.Request, args *struct{}, reply *APIWalletGetAccountsReply, authenticated bool) error {

	if !authenticated {
		return errors.New("Invalid User or Password")
	}

	api.wallet.RLock()
	defer api.wallet.RUnlock()

	reply.Version = api.wallet.Version
	reply.Encrypted = api.wallet.Encryption.Encrypted

	reply.Addresses = make([]*APIWalletReplyAddress, len(api.wallet.Addresses))
	for i, addr := range api.wallet.Addresses {
		reply.Addresses[i] = importWalletAddress(addr)
	}

	return nil
}

func (api *APICommon) WalletGetAddresses_http(values url.Values) (interface{}, error) {
	args := &APIWalletGetAccounts{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIWalletGetAccountsReply{}
	return reply, api.WalletGetAddresses(nil, &struct{}{}, reply, args.CheckAuthenticated())
}

func (api *APICommon) WalletGetAddresses_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &APIWalletGetAccountsReply{}
	return reply, api.WalletGetAddresses(nil, &struct{}{}, reply, conn.Authenticated.IsSet())
}
