package info

type AssetInfo struct {
	Version          uint64 `json:"version" msgpack:"version"`
	Name             string `json:"name" msgpack:"name"`
	Ticker           string `json:"ticker" msgpack:"ticker"`
	Identification   string `json:"identification" msgpack:"identification"`
	DecimalSeparator byte   `json:"decimalSeparator" msgpack:"decimalSeparator"`
	Description      string `json:"description,omitempty" msgpack:"description,omitempty"`
	Hash             []byte `json:"hash,omitempty" msgpack:"hash,omitempty"`
}
