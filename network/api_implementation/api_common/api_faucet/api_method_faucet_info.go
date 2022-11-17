package api_faucet

type APIFaucetInfo struct {
	FaucetTestnetEnabled bool   `json:"faucetTestnetEnabled,omitempty" msgpack:"faucetTestnetEnabled,omitempty"`
	Origin               string `json:"origin,omitempty" msgpack:"origin,omitempty"`
	ChallengeUri         string `json:"challengeUri,omitempty" msgpack:"challengeUri,omitempty"`
	FaucetTestnetCoins   uint64 `json:"faucetTestnetCoins,omitempty" msgpack:"faucetTestnetCoins,omitempty"`
}
