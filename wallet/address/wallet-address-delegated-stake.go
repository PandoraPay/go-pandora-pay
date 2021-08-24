package wallet_address

import (
	"pandora-pay/addresses"
	"pandora-pay/helpers"
)

type WalletAddressDelegatedStake struct {
	PrivateKey     *addresses.PrivateKey `json:"privateKey"`
	PublicKey      helpers.HexBytes      `json:"publicKey"`
	LastKnownNonce uint32                `json:"lastKnownNonce"`
}
