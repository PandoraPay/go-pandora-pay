package wallet_address

import (
	"pandora-pay/addresses"
	"pandora-pay/helpers"
)

type WalletAddressDelegatedStake struct {
	PrivateKey     *addresses.PrivateKey `json:"privateKey"`
	PublicKeyHash  helpers.HexBytes      `json:"publicKeyHash"`
	LastKnownNonce uint32                `json:"lastKnownNonce"`
}
