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
	Height uint64           `json:"height"`
	Hash   helpers.HexBytes `json:"hash"`
}

type APIBlockCompleteRequest struct {
	Height     uint64           `json:"height"`
	Hash       helpers.HexBytes `json:"hash"`
	ReturnType APIReturnType    `json:"returnType"`
}

type APITransactionRequest struct {
	Height     uint64           `json:"height"`
	Hash       helpers.HexBytes `json:"hash"`
	ReturnType APIReturnType    `json:"returnType"`
}

type APIAccountRequest struct {
	Address    string           `json:"address"`
	Hash       helpers.HexBytes `json:"hash"`
	ReturnType APIReturnType    `json:"returnType"`
}

func (request *APIAccountRequest) GetPublicKeyHash() ([]byte, error) {
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
