package wallet_address

import "pandora-pay/helpers"

type WalletAddressBalanceDecrypted struct {
	Amount  uint64           `json:"amount" msgpack:"amount"`
	Balance helpers.HexBytes `json:"balance" msgpack:"balance"`
}
