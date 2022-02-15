package wallet_address

import (
	"pandora-pay/addresses"
)

type WalletAddressDelegatedStake struct {
	PrivateKey     *addresses.PrivateKey `json:"privateKey" msgpack:"privateKey"`
	PublicKey      []byte                `json:"publicKey" msgpack:"publicKey"`
	LastKnownNonce uint32                `json:"lastKnownNonce" msgpack:"lastKnownNonce"`
}
