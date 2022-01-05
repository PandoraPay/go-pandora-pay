package api_types

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/config/config_auth"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
)

type SubscriptionType uint8

const (
	SUBSCRIPTION_ACCOUNT SubscriptionType = iota
	SUBSCRIPTION_PLAIN_ACCOUNT
	SUBSCRIPTION_ACCOUNT_TRANSACTIONS
	SUBSCRIPTION_ASSET
	SUBSCRIPTION_REGISTRATION
	SUBSCRIPTION_TRANSACTION
)

type APIReturnType uint8

const (
	RETURN_JSON APIReturnType = iota
	RETURN_SERIALIZED
)

type APIAccountBaseRequest struct {
	Address   string           `json:"address,omitempty"  urlstruct:"address"`
	PublicKey helpers.HexBytes `json:"publicKey,omitempty"  urlstruct:"publicKey"`
}

func (request *APIAccountBaseRequest) GetPublicKey() ([]byte, error) {
	var publicKey []byte
	if request.Address != "" {
		address, err := addresses.DecodeAddr(request.Address)
		if err != nil {
			return nil, errors.New("Invalid address")
		}
		publicKey = address.PublicKey
	} else if request.PublicKey != nil && len(request.PublicKey) == cryptography.PublicKeySize {
		publicKey = request.PublicKey
	} else {
		return nil, errors.New("Invalid address")
	}

	return publicKey, nil
}

type APISubscriptionRequest struct {
	Key        []byte           `json:"key,omitempty"`
	Type       SubscriptionType `json:"type,omitempty"`
	ReturnType APIReturnType    `json:"returnType,omitempty"`
}

type APIUnsubscriptionRequest struct {
	Key  []byte           `json:"key,omitempty"`
	Type SubscriptionType `json:"type,omitempty"`
}

type APIAuthenticateBaseRequest struct {
	Username string `json:"user" urlstruct:"user"`
	Password string `json:"pass" urlstruct:"pass"`
}

func (request *APIAuthenticateBaseRequest) CheckAuthenticated() bool {
	user := config_auth.CONFIG_AUTH_USERS_MAP[request.Username]
	if user == nil {
		return false
	}

	return user.Password == request.Password
}
