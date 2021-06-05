package api_common

import (
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
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

type APIBlockRequest struct {
	Height uint64           `json:"height,omitempty"`
	Hash   helpers.HexBytes `json:"hash,omitempty"`
}

type APIBlockCompleteRequest struct {
	Height     uint64           `json:"height,omitempty"`
	Hash       helpers.HexBytes `json:"hash,omitempty"`
	ReturnType APIReturnType    `json:"returnType,omitempty"`
}

type APITransactionRequest struct {
	Height     uint64           `json:"height,omitempty"`
	Hash       helpers.HexBytes `json:"hash,omitempty"`
	ReturnType APIReturnType    `json:"returnType,omitempty"`
}

type APIAccountRequestData struct {
	Address string           `json:"address,omitempty"`
	Hash    helpers.HexBytes `json:"hash,omitempty"`
}

type APIAccountRequest struct {
	APIAccountRequestData
	ReturnType APIReturnType `json:"returnType,omitempty"`
}

type APIAccountUnsubscribeRequest struct {
	APIAccountRequestData
}

func (request *APIAccountRequestData) GetPublicKeyHash() ([]byte, error) {
	var publicKeyHash []byte
	if request.Address != "" {
		address, err := addresses.DecodeAddr(request.Address)
		if err != nil {
			return nil, errors.New("Invalid address")
		}
		publicKeyHash = address.PublicKeyHash
	} else if request.Hash != nil && len(request.Hash) == cryptography.PublicKeyHashHashSize {
		publicKeyHash = request.Hash
	} else {
		return nil, errors.New("Invalid address")
	}

	return publicKeyHash, nil
}

type APITokenRequest struct {
	Hash       helpers.HexBytes `json:"hash"`
	ReturnType APIReturnType    `json:"returnType"`
}
