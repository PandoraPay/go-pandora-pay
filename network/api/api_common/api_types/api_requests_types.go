package api_types

import (
	"encoding/hex"
	"errors"
	"net/url"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"strconv"
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

func GetReturnType(s string, defaultValue APIReturnType) APIReturnType {
	switch s {
	case "0":
		return RETURN_JSON
	case "1":
		return RETURN_SERIALIZED
	default:
		return defaultValue
	}
}

type APIAccountBaseRequest struct {
	Address   string           `json:"address,omitempty"`
	PublicKey helpers.HexBytes `json:"publicKey,omitempty"`
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
	Key  []byte           `json:"Key,omitempty"`
	Type SubscriptionType `json:"type,omitempty"`
}

type APIHeightHash struct {
	Height uint64           `json:"height,omitempty"`
	Hash   helpers.HexBytes `json:"hash,omitempty"`
}

func (self *APIHeightHash) ImportFromValues(values *url.Values) (err error) {
	err = errors.New("parameter 'hash' or 'height' are missing")
	if values.Get("height") != "" {
		self.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		self.Hash, err = hex.DecodeString(values.Get("hash"))
	}
	return
}

func (self *APIAccountBaseRequest) ImportFromValues(values *url.Values) (err error) {

	if values.Get("address") != "" {
		self.Address = values.Get("address")
	} else if values.Get("publicKey") != "" {
		self.PublicKey, err = hex.DecodeString(values.Get("publicKey"))
	} else {
		err = errors.New("parameter 'address' or 'hash' was not specified")
	}

	return
}
