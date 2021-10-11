package info

type AssetInfo struct {
	Name             string `json:"name"`
	Ticker           string `json:"ticker"`
	DecimalSeparator byte   `json:"decimalSeparator"`
	Description      string `json:"description,omitempty"`
}
