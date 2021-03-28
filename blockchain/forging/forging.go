package forging

import (
	"github.com/tevino/abool"
	"math/big"
	"pandora-pay/blockchain/block-complete"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/mempool"
)

type Forging struct {
	mempool         *mempool.Mempool
	Wallet          *ForgingWallet
	workChannel     chan *ForgingWork
	started         *abool.AtomicBool
	SolutionChannel chan *block_complete.BlockComplete
}

func ForgingInit(mempool *mempool.Mempool) (forging *Forging, err error) {

	forging = &Forging{
		mempool:         mempool,
		workChannel:     nil,
		started:         abool.New(),
		SolutionChannel: make(chan *block_complete.BlockComplete),
		Wallet: &ForgingWallet{
			addressesMap: make(map[string]*ForgingWalletAddress),
		},
	}

	gui.Log("Forging Init")

	return
}

func (forging *Forging) StartForging() bool {

	if !forging.started.SetToIf(false, true) {
		return false
	}

	forging.workChannel = make(chan *ForgingWork)
	forgingThread := createForgingThread(config.CPU_THREADS, forging.mempool, forging.SolutionChannel, forging.workChannel, forging.Wallet)
	go forgingThread.startForging()

	return true
}

func (forging *Forging) StopForging() bool {
	if forging.started.SetToIf(true, false) {
		close(forging.workChannel) //this will close the thread
		return true
	}
	return false
}

//thread safe
func (forging *Forging) ForgingNewWork(blkComplete *block_complete.BlockComplete, target *big.Int) {

	work := &ForgingWork{
		blkComplete: blkComplete,
		target:      target,
	}

	if forging.started.IsSet() {
		forging.workChannel <- work
	}
}

func (forging *Forging) Close() {
	forging.StopForging()
	close(forging.SolutionChannel) //this will close the thread
}
