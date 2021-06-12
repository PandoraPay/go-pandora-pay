package forging

import (
	"github.com/tevino/abool"
	"pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/forging/forging-block-work"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
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
	solutionCn         chan<- *block_complete.BlockComplete
}

func CreateForging(mempool *mempool.Mempool) (forging *Forging, err error) {

	forging = &Forging{
		mempool,
		&ForgingWallet{
			[]*ForgingWalletAddress{},
			map[string]*ForgingWalletAddress{},
			[]int{},
			[]*ForgingWorkerThread{},
			nil,
			make(chan *ForgingWalletAddressUpdate),
			make(chan struct{}),
			nil,
			nil,
			nil,
		},
		abool.New(),
		nil, nil, nil,
	}
	forging.Wallet.forging = forging

	return
}

func (forging *Forging) InitializeForging(nextBlockCreatedCn <-chan *forging_block_work.ForgingWork, updateAccounts *multicast.MulticastChannel, forgingSolutionCn chan<- *block_complete.BlockComplete) {

	forging.nextBlockCreatedCn = nextBlockCreatedCn
	forging.Wallet.updateAccounts = updateAccounts
	forging.solutionCn = forgingSolutionCn

	forging.forgingThread = createForgingThread(config.CPU_THREADS, forging.mempool, forging.solutionCn, forging.nextBlockCreatedCn)
	forging.Wallet.workersCreatedCn = forging.forgingThread.workersCreatedCn
	forging.Wallet.workersDestroyedCn = forging.forgingThread.workersDestroyedCn

	recovery.SafeGo(forging.Wallet.processUpdates)

}

func (forging *Forging) StartForging() bool {

	if globals.Arguments["--staking"] == false {
		gui.GUI.Warning(`Staking was not started as "--staking" is missing`)
		return false
	}
	if config.CONSENSUS != config.CONSENSUS_TYPE_FULL {
		gui.GUI.Warning(`Staking was not started as "--consensus=full" is missing`)
		return false
	}

	if !forging.started.SetToIf(false, true) {
		return false
	}

	recovery.SafeGo(forging.forgingThread.startForging)

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
