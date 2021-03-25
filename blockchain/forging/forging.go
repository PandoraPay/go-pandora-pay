package forging

import (
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
	SolutionChannel chan *block_complete.BlockComplete
}

func ForgingInit(mempool *mempool.Mempool) (forging *Forging, err error) {

	forging = &Forging{
		mempool:         mempool,
		workChannel:     make(chan *ForgingWork),
		SolutionChannel: make(chan *block_complete.BlockComplete),
		Wallet: &ForgingWallet{
			addressesMap: make(map[string]*ForgingWalletAddress),
		},
	}

	gui.Log("Forging Init")
	if err = forging.Wallet.loadBalances(); err != nil {
		return
	}

	forgingThread := createForgingThread(config.CPU_THREADS, mempool, forging.SolutionChannel, forging.workChannel, forging.Wallet)
	go forgingThread.startForging()

	return
}

func (forging *Forging) StopForging() {
	close(forging.workChannel)     //this will close the thread
	close(forging.SolutionChannel) //this will close the thread
}

//thread safe
func (forging *Forging) ForgingNewWork(blkComplete *block_complete.BlockComplete, target *big.Int) {

	work := &ForgingWork{
		blkComplete: blkComplete,
		target:      target,
	}

	forging.workChannel <- work
}

func (forging *Forging) Close() {
	forging.StopForging()
}
