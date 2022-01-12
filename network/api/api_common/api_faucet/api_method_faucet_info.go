package api_faucet

type APIFaucetInfo struct {
	HCaptchaSiteKey      string `json:"hCaptchaSiteKey,omitempty" msgpack:"hCaptchaSiteKey,omitempty"`
	FaucetTestnetEnabled bool   `json:"faucetTestnetEnabled,omitempty" msgpack:"faucetTestnetEnabled,omitempty"`
	FaucetTestnetCoins   uint64 `json:"faucetTestnetCoins,omitempty" msgpack:"faucetTestnetCoins,omitempty"`
}
