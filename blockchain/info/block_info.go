package info

import "pandora-pay/helpers"

type BlockInfo struct {
	Hash       helpers.HexBytes `json:"hash"`       //32 bytes
	KernelHash helpers.HexBytes `json:"kernelHash"` //32 bytes
	Timestamp  uint64           `json:"timestamp"`
	Size       uint64           `json:"size"`
	TXs        uint64           `json:"txs"`
	Forger     helpers.HexBytes `json:"forger"` //20 bytes
}
