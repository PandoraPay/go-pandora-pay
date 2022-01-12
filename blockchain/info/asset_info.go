package info

import "pandora-pay/helpers"

type AssetInfo struct {
	Version          uint64           `json:"version" msgpack:"version"`
	Name             string           `json:"name" msgpack:"name"`
	Ticker           string           `json:"ticker" msgpack:"ticker"`
	DecimalSeparator byte             `json:"decimalSeparator" msgpack:"decimalSeparator"`
	Description      string           `json:"description,omitempty" msgpack:"description,omitempty"`
	Hash             helpers.HexBytes `json:"hash,omitempty" msgpack:"hash,omitempty"`
}
