package api_common

import (
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
	Address    string           `json:"height"`
	Hash       helpers.HexBytes `json:"hash"`
	ReturnType APIReturnType    `json:"returnType"`
}

type APITokenRequest struct {
	Hash       helpers.HexBytes `json:"hash"`
	ReturnType APIReturnType    `json:"returnType"`
}
