package wallet

import (
	"pandora-pay/addresses"
	"pandora-pay/helpers"
)

type WalletAddress struct {
	Name           string
	PrivateKey     *addresses.PrivateKey
	PublicKey      helpers.ByteString //33 byte
	PublicKeyHash  helpers.ByteString //20 byte
	AddressEncoded string
	Address        *addresses.Address
	SeedIndex      uint32
}
