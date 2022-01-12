package info

import "pandora-pay/helpers"

type BlockInfo struct {
	Hash       helpers.HexBytes `json:"hash" msgpack:"hash"`              //32 bytes
	KernelHash helpers.HexBytes `json:"kernelHash"  msgpack:"kernelHash"` //32 bytes
	Timestamp  uint64           `json:"timestamp"  msgpack:"timestamp"`
	Size       uint64           `json:"size"  msgpack:"size"`
	TXs        uint64           `json:"txs"  msgpack:"txs"`
	Fees       uint64           `json:"fees"  msgpack:"fees"`
	Forger     helpers.HexBytes `json:"forger"  msgpack:"forger"` //20 bytes
}
