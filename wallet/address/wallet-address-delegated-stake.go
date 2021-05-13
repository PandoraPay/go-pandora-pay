package wallet_address

import (
	"pandora-pay/addresses"
	"pandora-pay/helpers"
)

type WalletAddressDelegatedStake struct {
	PrivateKey     *addresses.PrivateKey
	PublicKeyHash  helpers.HexBytes
	LastKnownNonce uint32
}
