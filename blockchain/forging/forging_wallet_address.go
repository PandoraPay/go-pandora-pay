package forging

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/helpers"
)

type ForgingWalletAddress struct {
	delegatedPrivateKey     *addresses.PrivateKey
	delegatedStakePublicKey helpers.HexBytes //20 byte
	delegatedStakeFee       uint64
	publicKey               helpers.HexBytes //20byte
	publicKeyStr            string
	plainAcc                *plain_account.PlainAccount
	workerIndex             int
}

func (walletAddr *ForgingWalletAddress) clone() *ForgingWalletAddress {
	return &ForgingWalletAddress{
		walletAddr.delegatedPrivateKey,
		walletAddr.delegatedStakePublicKey,
		walletAddr.delegatedStakeFee,
		walletAddr.publicKey,
		walletAddr.publicKeyStr,
		walletAddr.plainAcc,
		walletAddr.workerIndex,
	}
}
