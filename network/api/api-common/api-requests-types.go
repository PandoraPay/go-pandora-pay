package api_common

import "pandora-pay/helpers"

type APIReturnType uint8

const (
	RETURN_JSON APIReturnType = iota
	RETURN_SERIALIZED
)

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
