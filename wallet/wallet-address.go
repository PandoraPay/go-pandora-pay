package wallet

import "pandora-pay/addresses"

type WalletAddress struct {
	Name       string
	PrivateKey *addresses.PrivateKey
	PublicKey  []byte
	Address    *addresses.Address
	SeedIndex  uint32
}
