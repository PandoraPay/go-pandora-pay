package info

type BlockInfo struct {
	Hash       []byte `json:"hash" msgpack:"hash"`              //32 bytes
	KernelHash []byte `json:"kernelHash"  msgpack:"kernelHash"` //32 bytes
	Timestamp  uint64 `json:"timestamp"  msgpack:"timestamp"`
	Size       uint64 `json:"size"  msgpack:"size"`
	TXs        uint64 `json:"txs"  msgpack:"txs"`
	Fees       uint64 `json:"fees"  msgpack:"fees"`
}
