package shared_staked

import (
	"pandora-pay/addresses"
)

type WalletAddressSharedStaked struct {
	PrivateKey *addresses.PrivateKey `json:"privateKey" msgpack:"privateKey"`
	PublicKey  []byte                `json:"publicKey" msgpack:"publicKey"`
}
