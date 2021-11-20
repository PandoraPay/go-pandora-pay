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

type apiPendingDelegateStakeChange struct {
	delegateStakingPrivateKey *addresses.PrivateKey
	delegateStakingPublicKey  []byte
	publicKey                 []byte
	blockHeight               uint64
}

type APIDelegatorNode struct {
	challenge                     []byte
	chainHeight                   uint64    //use atomic
	pendingDelegatesStakesChanges *sync.Map //*apiPendingDelegateStakeChange
	ticker                        *time.Ticker
	wallet                        *wallet.Wallet
	chain                         *blockchain.Blockchain
}

func CreateDelegatorNode(chain *blockchain.Blockchain, wallet *wallet.Wallet) (delegator *APIDelegatorNode) {

	challenge := helpers.RandomBytes(cryptography.HashSize)

	delegator = &APIDelegatorNode{
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
