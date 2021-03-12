package wallet

import "pandora-pay/addresses"

type WalletAddress struct {
	Name           string
	PrivateKey     *addresses.PrivateKey
	PublicKey      []byte //33 byte
	PublicKeyHash  []byte //20 byte
	AddressEncoded string
	Address        *addresses.Address
	SeedIndex      uint32
}
