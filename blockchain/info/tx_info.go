package info

type TxInfo struct {
	Height    uint64 `json:"height" msgpack:"height"`
	BlkHeight uint64 `json:"blkHeight" msgpack:"blkHeight"`
	Timestmap uint64 `json:"timestamp" msgpack:"timestamp"`
}
