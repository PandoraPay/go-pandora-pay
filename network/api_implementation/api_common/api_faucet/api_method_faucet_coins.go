package api_faucet

type APIFaucetCoinsRequest struct {
	Address     string `json:"address,omitempty" msgpack:"address,omitempty"`
	FaucetToken string `json:"faucetToken,omitempty" msgpack:"faucetToken,omitempty"`
}

type APIFaucetCoinsReply struct {
	Hash []byte `json:"hash" msgpack:"hash"`
}
