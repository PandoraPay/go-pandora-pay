package consensus

import (
	"pandora-pay/blockchain"
	"pandora-pay/helpers/generics"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
	"pandora-pay/txs_validator"
)

type Consensus struct {
	chain        *blockchain.Blockchain
	txsValidator *txs_validator.TxsValidator
	mempool      *mempool.Mempool
	forks        *Forks
}

func (consensus *Consensus) execute() {
	//discover forks
	processForksThread := newConsensusProcessForksThread(consensus.forks, consensus.txsValidator, consensus.chain, consensus.mempool)
	recovery.SafeGo(processForksThread.execute)
}

func NewConsensus(chain *blockchain.Blockchain, mempool *mempool.Mempool, txsValidator *txs_validator.TxsValidator) *Consensus {

	consensus := &Consensus{
		chain,
		txsValidator,
		mempool,
		&Forks{
			hashes: &generics.Map[string, *Fork]{},
		},
	}

	consensus.execute()

	return consensus
}
