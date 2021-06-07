package token_info

type TokenInfo struct {
	Name             string `json:"name"`
	Ticker           string `json:"ticker"`
	DecimalSeparator byte   `json:"decimalSeparator"`
	Description      string `json:"description,omitempty"`
}
