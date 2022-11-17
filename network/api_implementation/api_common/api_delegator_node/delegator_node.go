package api_delegator_node

import (
	"pandora-pay/blockchain"
	"pandora-pay/wallet"
)

type DelegatorNode struct {
	chainHeight uint64 //use atomic
	wallet      *wallet.Wallet
	chain       *blockchain.Blockchain
}

func NewDelegatorNode(chain *blockchain.Blockchain, wallet *wallet.Wallet) (delegator *DelegatorNode) {

	delegator = &DelegatorNode{
		0,
		wallet,
		chain,
	}

	return
}
