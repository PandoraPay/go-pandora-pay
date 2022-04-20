package forging

import (
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
)

type ForgingWalletAddress struct {
	publicKeyHash            []byte
	publicKeyHashStr         string
	delegatedStakePrivateKey []byte
	delegatedStakePublicKey  []byte
	plainAcc                 *plain_account.PlainAccount
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
		walletAddr.plainAcc,
		walletAddr.stakingAvailable,
		walletAddr.workerIndex,
		walletAddr.chainHash,
	}
}
