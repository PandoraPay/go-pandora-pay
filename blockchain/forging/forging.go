package forging

import (
	"github.com/tevino/abool"
	"pandora-pay/blockchain/block-complete"
	forging_block_work "pandora-pay/blockchain/forging/forging-block-work"
	"pandora-pay/config"
	"pandora-pay/config/globals"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"sync"
	"sync/atomic"
)

type Forging struct {
	mempool            *mempool.Mempool
	Wallet             *ForgingWallet
	started            *abool.AtomicBool
	workCn             chan *forging_block_work.ForgingWork
	nextBlockCreatedCn <-chan *forging_block_work.ForgingWork
	solutionCn         chan<- *block_complete.BlockComplete
}

func ForgingInit(mempool *mempool.Mempool, nextBlockCreated <-chan *forging_block_work.ForgingWork, updateAccounts *multicast.MulticastChannel, solutionCn chan<- *block_complete.BlockComplete) (forging *Forging, err error) {

	forging = &Forging{
		mempool:            mempool,
		workCn:             nil,
		started:            abool.New(),
		solutionCn:         solutionCn,
		nextBlockCreatedCn: nextBlockCreated,
		Wallet: &ForgingWallet{
			addressesMap:   make(map[string]*ForgingWalletAddress),
			updates:        &atomic.Value{},
			updatesMutex:   &sync.Mutex{},
			updateAccounts: updateAccounts,
		},
	}

	forging.Wallet.updates.Store([]*ForgingWalletAddressUpdate{})

	if config.CONSENSUS == config.CONSENSUS_TYPE_FULL {
		go forging.Wallet.updateAccountsChanges()
		go forging.forgingNewWork()
	}

	gui.GUI.Log("Forging Init")

	return
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

	if err := forging.Wallet.ProcessUpdates(); err != nil {
		forging.started.UnSet()
		return false
	}

	forging.workCn = make(chan *forging_block_work.ForgingWork, 10)
	forgingThread := createForgingThread(config.CPU_THREADS, forging.mempool, forging.solutionCn, forging.workCn, forging.Wallet)

	go forgingThread.startForging()

	return true
}

func (forging *Forging) StopForging() bool {
	if forging.started.SetToIf(true, false) {
		close(forging.workCn) //this will close the thread
		return true
	}
	return false
}

//thread safe
func (forging *Forging) forgingNewWork() {

	for {

		work, ok := <-forging.nextBlockCreatedCn
		if !ok {
			return
		}

		if forging.started.IsSet() {
			forging.workCn <- work
		}

	}

}

func (forging *Forging) Close() {
	forging.StopForging()
}
