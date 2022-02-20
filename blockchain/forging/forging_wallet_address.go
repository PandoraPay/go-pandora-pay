package forging

import (
	"math/big"
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
)

type ForgingWalletAddress struct {
	privateKey      *addresses.PrivateKey
	privateKeyPoint *big.Int
	publicKey       []byte
	publicKeyStr    string
	plainAcc        *plain_account.PlainAccount
	workerIndex     int
}

func (walletAddr *ForgingWalletAddress) clone() *ForgingWalletAddress {
	return &ForgingWalletAddress{
		walletAddr.privateKey,
		walletAddr.privateKeyPoint,
		walletAddr.publicKey,
		walletAddr.publicKeyStr,
		walletAddr.plainAcc,
		walletAddr.workerIndex,
	}
}
