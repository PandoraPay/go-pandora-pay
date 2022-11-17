package api_code_types

import (
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

type APISubscriptionNotification struct {
	SubscriptionType SubscriptionType `json:"type,omitempty" msgpack:"type,omitempty"`
	Key              []byte           `json:"key,omitempty" msgpack:"key,omitempty"`
	Data             []byte           `json:"data,omitempty" msgpack:"data,omitempty"`
	Extra            []byte           `json:"extra,omitempty" msgpack:"extra,omitempty"`
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
