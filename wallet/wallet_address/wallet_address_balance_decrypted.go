package wallet_address

import "pandora-pay/helpers"

type WalletAddressDecryptedBalance struct {
	Amount           uint64           `json:"amount" msgpack:"amount"`
	EncryptedBalance helpers.HexBytes `json:"encryptedBalance" msgpack:"encryptedBalance"`
}
