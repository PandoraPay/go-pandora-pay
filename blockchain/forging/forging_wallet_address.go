package forging

import (
	"math/big"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/accounts/account"
)

type ForgingWalletAddress struct {
	privateKey              *addresses.PrivateKey
	privateKeyPoint         *big.Int
	publicKey               []byte
	publicKeyStr            string
	account                 *account.Account
	decryptedStakingBalance uint64
	workerIndex             int
}

func (walletAddr *ForgingWalletAddress) clone() *ForgingWalletAddress {
	return &ForgingWalletAddress{
		walletAddr.privateKey,
		walletAddr.privateKeyPoint,
		walletAddr.publicKey,
		walletAddr.publicKeyStr,
		walletAddr.account,
		walletAddr.decryptedStakingBalance,
		walletAddr.workerIndex,
	}
}
