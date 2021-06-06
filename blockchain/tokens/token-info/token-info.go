package token_info

import "pandora-pay/helpers"

type TokenInfo struct {
	Hash             helpers.HexBytes `json:"hash"` //20 bytes
	Name             string           `json:"name"`
	Ticker           string           `json:"ticker"`
	DecimalSeparator byte             `json:"decimalSeparator"`
	Description      string           `json:"description,omitempty"`
}
