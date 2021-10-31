package info

type AssetInfo struct {
	Version          uint64 `json:"version"`
	Name             string `json:"name"`
	Ticker           string `json:"ticker"`
	DecimalSeparator byte   `json:"decimalSeparator"`
	Description      string `json:"description,omitempty"`
}
