package wallet_address

import "pandora-pay/addresses"

type WalletAddressDelegatedStake struct {
	PrivateKey     *addresses.PrivateKey
	PublicKeyHash  []byte
	LastKnownNonce uint32
}
