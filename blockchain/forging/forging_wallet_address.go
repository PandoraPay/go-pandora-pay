package forging

import (
	"pandora-pay/addresses"
)

type ForgingWalletAddress struct {
	publicKeyHash            []byte
	publicKeyHashStr         string
	delegatedStakePrivateKey *addresses.PrivateKey
	delegatedStakePublicKey  []byte
	delegatedStakeFee        uint64
	stakingAvailable         uint64
	workerIndex              int
	chainHash                []byte
}

func (walletAddr *ForgingWalletAddress) clone() *ForgingWalletAddress {
	return &ForgingWalletAddress{
		walletAddr.publicKeyHash,
		walletAddr.publicKeyHashStr,
		walletAddr.delegatedStakePrivateKey,
		walletAddr.delegatedStakePublicKey,
		walletAddr.delegatedStakeFee,
		walletAddr.stakingAvailable,
		walletAddr.workerIndex,
		walletAddr.chainHash,
	}
}
