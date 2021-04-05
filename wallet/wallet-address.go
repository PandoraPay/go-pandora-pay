package wallet

import (
	"pandora-pay/addresses"
	"pandora-pay/helpers"
)

type WalletAddress struct {
	Name           string
	PrivateKey     *addresses.PrivateKey
	PublicKey      helpers.HexBytes //33 byte
	PublicKeyHash  helpers.HexBytes //20 byte
	AddressEncoded string
	Address        *addresses.Address
	SeedIndex      uint32
}
