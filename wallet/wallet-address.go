package wallet

import "pandora-pay/addresses"

type WalletAddressVersion int

const (
	WalletAddressTransparent WalletAddressVersion = 0
)

func (e WalletAddressVersion) String() string {
	switch e {
	case WalletAddressTransparent:
		return "Transparent"
	default:
		return "Unknown Wallet Address Version"
	}
}

type WalletAddress struct {
	Version         WalletAddressVersion
	Name            string
	PrivateKey      *addresses.PrivateKey
	Address         *addresses.Address
	IsSeedGenerated bool
}
