package api_types

import (
	"errors"
	"net/url"
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
	SUBSCRIPTION_TRANSACTION
)

type APIReturnType uint8

const (
	RETURN_SERIALIZED APIReturnType = iota
	RETURN_JSON
)

type APIAccountBaseRequest struct {
	Address       string         `json:"address,omitempty" msgpack:"address,omitempty"`
	PublicKeyHash helpers.Base64 `json:"publicKeyHash,omitempty"  msgpack:"publicKeyHash,omitempty"`
}

func (request *APIAccountBaseRequest) GetPublicKeyHash(required bool) ([]byte, error) {
	if request == nil {
		return nil, errors.New("argument missing")
	}

	var publicKeyHash []byte
	if request.Address != "" {
		address, err := addresses.DecodeAddr(request.Address)
		if err != nil {
			return nil, errors.New("Invalid address")
		}
		publicKeyHash = address.PublicKeyHash
	} else if request.PublicKeyHash != nil && len(request.PublicKeyHash) == cryptography.PublicKeyHashSize {
		publicKeyHash = request.PublicKeyHash
	} else if required {
		return nil, errors.New("Invalid address or publicKey")
	}

	return publicKeyHash, nil
}

type APISubscriptionRequest struct {
	Key        helpers.Base64   `json:"key,omitempty" msgpack:"key,omitempty"`
	Type       SubscriptionType `json:"type,omitempty"  msgpack:"type,omitempty"`
	ReturnType APIReturnType    `json:"returnType,omitempty"  msgpack:"returnType,omitempty"`
}

type APIUnsubscriptionRequest struct {
	Key  helpers.Base64   `json:"key,omitempty" msgpack:"key,omitempty"`
	Type SubscriptionType `json:"type,omitempty" msgpack:"type,omitempty"`
}

type APIAuthenticated[T any] struct {
	User string `json:"user" msgpack:"user"`
	Pass string `json:"pass" msgpack:"pass"`
	Data *T     `json:"req" msgpack:"req"`
}

func CheckAuthenticated(args url.Values) bool {

	user := config_auth.CONFIG_AUTH_USERS_MAP[args.Get("user")]
	if user == nil {
		return false
	}

	return user.Password == args.Get("pass")
}

func (authenticated *APIAuthenticated[T]) CheckAuthenticated() bool {
	user := config_auth.CONFIG_AUTH_USERS_MAP[authenticated.User]
	if user == nil {
		return false
	}

	return user.Password == authenticated.Pass
}
