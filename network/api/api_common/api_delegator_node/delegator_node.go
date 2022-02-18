package api_delegator_node

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/wallet"
)

type PendingDelegateStakeChange struct {
	delegateStakingPrivateKey *addresses.PrivateKey
	delegateStakingPublicKey  []byte
	publicKey                 []byte
	blockHeight               uint64
}

type DelegatorNode struct {
	challenge   []byte
	secret      []byte
	chainHeight uint64 //use atomic
	wallet      *wallet.Wallet
	chain       *blockchain.Blockchain
}

func NewDelegatorNode(chain *blockchain.Blockchain, wallet *wallet.Wallet) (delegator *DelegatorNode) {

	delegator = &DelegatorNode{
		helpers.RandomBytes(cryptography.HashSize),
		helpers.RandomBytes(cryptography.HashSize),
		0,
		wallet,
		chain,
	}

	return
}
