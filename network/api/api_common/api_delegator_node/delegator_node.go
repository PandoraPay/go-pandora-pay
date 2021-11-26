package api_delegator_node

import (
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/wallet"
	"sync"
	"time"
)

type pendingDelegateStakeChange struct {
	delegateStakingPrivateKey *addresses.PrivateKey
	delegateStakingPublicKey  []byte
	publicKey                 []byte
	blockHeight               uint64
}

type DelegatorNode struct {
	challenge                     []byte
	chainHeight                   uint64    //use atomic
	pendingDelegatesStakesChanges *sync.Map //*pendingDelegateStakeChange
	ticker                        *time.Ticker
	wallet                        *wallet.Wallet
	chain                         *blockchain.Blockchain
}

func NewDelegatorNode(chain *blockchain.Blockchain, wallet *wallet.Wallet) (delegator *DelegatorNode) {

	challenge := helpers.RandomBytes(cryptography.HashSize)

	delegator = &DelegatorNode{
		challenge,
		0,
		&sync.Map{},
		nil,
		wallet,
		chain,
	}

	delegator.execute()

	return
}
