package api_delegator_node

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"pandora-pay/wallet"
)

type PendingDelegateStakeChange struct {
	delegateStakingPrivateKey *addresses.PrivateKey
	delegateStakingPublicKey  []byte
	publicKey                 []byte
	blockHeight               uint64
}

type DelegatorNode struct {
	challenge                     []byte
	chainHeight                   uint64 //use atomic
	pendingDelegatesStakesChanges *generics.Map[string, *PendingDelegateStakeChange]
	wallet                        *wallet.Wallet
	chain                         *blockchain.Blockchain
}

func NewDelegatorNode(chain *blockchain.Blockchain, wallet *wallet.Wallet) (delegator *DelegatorNode) {

	challenge := helpers.RandomBytes(cryptography.HashSize)

	delegator = &DelegatorNode{
		challenge,
		0,
		&generics.Map[string, *PendingDelegateStakeChange]{},
		wallet,
		chain,
	}

	delegator.execute()

	return
}
