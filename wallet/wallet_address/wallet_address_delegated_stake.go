package wallet_address

import (
	"pandora-pay/addresses"
)

type WalletAddressSharedStake struct {
	PrivateKey *addresses.PrivateKey `json:"privateKey" msgpack:"privateKey"`
	PublicKey  []byte                `json:"publicKey" msgpack:"publicKey"`
}
