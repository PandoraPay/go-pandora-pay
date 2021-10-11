package wallet_address

import "pandora-pay/helpers"

type WalletAddressBalanceDecoded struct {
	AmountDecoded uint64           `json:"amount"`
	Token         helpers.HexBytes `json:"token"`
}
