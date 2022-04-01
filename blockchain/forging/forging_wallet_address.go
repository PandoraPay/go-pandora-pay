package forging

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/accounts/account"
)

type ForgingWalletAddress struct {
	privateKey              *addresses.PrivateKey
	publicKey               []byte
	publicKeyStr            string
	account                 *account.Account
	decryptedStakingBalance uint64
	workerIndex             int
	chainHash               []byte
}

func (walletAddr *ForgingWalletAddress) clone() *ForgingWalletAddress {
	return &ForgingWalletAddress{
		walletAddr.privateKey,
		walletAddr.publicKey,
		walletAddr.publicKeyStr,
		walletAddr.account,
		walletAddr.decryptedStakingBalance,
		walletAddr.workerIndex,
		walletAddr.chainHash,
	}
}
