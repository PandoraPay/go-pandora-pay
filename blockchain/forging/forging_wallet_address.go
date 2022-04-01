package forging

import (
	"pandora-pay/blockchain/data_storage/accounts/account"
)

type ForgingWalletAddress struct {
	publicKeyHash            []byte
	publicKeyHashStr         string
	delegatedStakePrivateKey []byte
	delegatedStakePublicKey  []byte
	account                  *account.Account
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
		walletAddr.account,
		walletAddr.stakingAvailable,
		walletAddr.workerIndex,
		walletAddr.chainHash,
	}
}
