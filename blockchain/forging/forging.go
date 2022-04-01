package forging

import (
	"github.com/tevino/abool"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/forging/forging_block_work"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/recovery"
)

type Forging struct {
	mempool            *mempool.Mempool
	Wallet             *ForgingWallet
	started            *abool.AtomicBool
	forgingThread      *ForgingThread
	nextBlockCreatedCn <-chan *forging_block_work.ForgingWork
	forgingSolutionCn  chan<- *blockchain_types.BlockchainSolution
}

func CreateForging(mempool *mempool.Mempool) (*Forging, error) {

	forging := &Forging{
		mempool,
		&ForgingWallet{
			map[string]*ForgingWalletAddress{},
			[]int{},
			[]*ForgingWorkerThread{},
			nil,
			make(chan *ForgingWalletAddressUpdate),
			nil,
			nil,
			&generics.Map[string, *ForgingWalletAddress]{},
			nil,
			abool.New(),
		},
		abool.New(),
		nil, nil, nil,
	}
	forging.Wallet.forging = forging

	return forging, nil
}

func (forging *Forging) InitializeForging(nextBlockCreatedCn <-chan *forging_block_work.ForgingWork, updateNewChainUpdate *multicast.MulticastChannel[*blockchain_types.BlockchainUpdates], forgingSolutionCn chan<- *blockchain_types.BlockchainSolution) {

	forging.nextBlockCreatedCn = nextBlockCreatedCn
	forging.Wallet.updateNewChainUpdate = updateNewChainUpdate
	forging.forgingSolutionCn = forgingSolutionCn

	forging.forgingThread = createForgingThread(config.CPU_THREADS, forging.mempool, forging.forgingSolutionCn, forging.nextBlockCreatedCn)
	forging.Wallet.workersCreatedCn = forging.forgingThread.workersCreatedCn
	forging.Wallet.workersDestroyedCn = forging.forgingThread.workersDestroyedCn

	forging.Wallet.initialized.Set()
	recovery.SafeGo(forging.Wallet.runProcessUpdates)
	recovery.SafeGo(forging.Wallet.runDecryptBalanceAndNotifyWorkers)

}

func (forging *Forging) StartForging() bool {

	if config.CONSENSUS != config.CONSENSUS_TYPE_FULL {
		gui.GUI.Warning(`Staking was not started as "--consensus=full" is missing`)
		return false
	}

	if !forging.started.SetToIf(false, true) {
		return false
	}

	forging.forgingThread.startForging()

	return true
}

func (forging *Forging) StopForging() bool {
	if forging.started.SetToIf(true, false) {
		return true
	}
	return false
}

func (forging *Forging) Close() {
	forging.StopForging()
}
