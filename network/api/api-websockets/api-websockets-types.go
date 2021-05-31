package api_websockets

import "pandora-pay/helpers"

type APIBlockHeight uint64

type APIBlockRequest struct {
	Height uint64
	Hash   helpers.HexBytes
}

type APIBlockCompleteRequest struct {
	Height     uint64
	Hash       helpers.HexBytes
	ReturnType uint8
}
