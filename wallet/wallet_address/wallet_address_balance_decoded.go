package wallet_address

import "pandora-pay/helpers"

type WalletAddressBalanceDecoded struct {
	AmountDecoded uint64           `json:"amount" msgpack:"amount"`
	Asset         helpers.HexBytes `json:"asset" msgpack:"asset"`
}
