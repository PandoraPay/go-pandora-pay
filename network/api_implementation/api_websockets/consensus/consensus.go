package consensus

import (
	"pandora-pay/blockchain"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/recovery"
	"pandora-pay/mempool"
)

type Consensus struct {
	chain   *blockchain.Blockchain
	mempool *mempool.Mempool
	forks   *Forks
}

func (consensus *Consensus) execute() {
	//discover forks
	processForksThread := newConsensusProcessForksThread(consensus.forks, consensus.chain, consensus.mempool)
	recovery.SafeGo(processForksThread.execute)
}

func NewConsensus(chain *blockchain.Blockchain, mempool *mempool.Mempool) *Consensus {

	consensus := &Consensus{
		chain,
		mempool,
		&Forks{
			hashes: &generics.Map[string, *Fork]{},
		},
	}

	consensus.execute()

	return consensus
}
