package consensus

import (
	"pandora-pay/blockchain"
	"pandora-pay/helpers/generics"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
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
		chain:   chain,
		mempool: mempool,
		forks: &Forks{
			hashes: &generics.Map[string, *Fork]{},
		},
	}

	consensus.execute()

	return consensus
}
