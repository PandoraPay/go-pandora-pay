package forging

import (
	"math/big"
	"pandora-pay/blockchain/block"
	"pandora-pay/config"
	"pandora-pay/gui"
	"sync"
	"sync/atomic"
	"time"
)

var started int32
var forgingWorking int32

type Forging struct {
	blkComplete *block.BlockComplete
	target      *big.Int

	solution          bool
	solutionTimestamp uint64
	solutionAddress   *ForgingWalletAddress

	Wallet ForgingWallet

	SolutionChannel chan *block.BlockComplete
}

var wg = sync.WaitGroup{}

func ForgingInit() (forging *Forging, err error) {

	forging = &Forging{
		SolutionChannel: make(chan *block.BlockComplete),
		Wallet: ForgingWallet{
			addressesMap: make(map[string]*ForgingWalletAddress),
		},
	}

	gui.Log("Forging Init")
	if err = forging.Wallet.loadBalances(); err != nil {
		return
	}

	go forging.startForging(config.CPU_THREADS)

	return
}

func (forging *Forging) startForging(threads int) {

	if !atomic.CompareAndSwapInt32(&started, 0, 1) {
		return
	}

	for atomic.LoadInt32(&started) == 1 {

		if forging.solution || forging.blkComplete == nil {
			// gui.Error("No block for staking..." )
			time.Sleep(10 * time.Millisecond)
			continue
		}

		if !atomic.CompareAndSwapInt32(&forgingWorking, 0, 1) {
			gui.Error("A strange error as forgingWorking couldn't be set to 1 ")
			return
		}

		for i := 0; i < threads; i++ {
			wg.Add(1)
			go forge(forging, threads, i)
		}

		wg.Wait()

		if forging.solution {
			err := forging.publishSolution()
			if err != nil {
				forging.solution = false
			}
		}

	}

}

func StopForging() {
	StopForgingWorkers()
	atomic.AddInt32(&started, -1)
}

func StopForgingWorkers() {
	atomic.CompareAndSwapInt32(&forgingWorking, 1, 0)
}

//thread safe
func (forging *Forging) RestartForgingWorkers(blkComplete *block.BlockComplete, target *big.Int) {

	atomic.CompareAndSwapInt32(&forgingWorking, 1, 0)

	wg.Wait()

	forging.solution = false
	forging.blkComplete = blkComplete
	forging.target = target

}

//thread safe
func (forging *Forging) foundSolution(address *ForgingWalletAddress, timestamp uint64) {

	if atomic.CompareAndSwapInt32(&forgingWorking, 1, 0) {
		forging.solution = true
		forging.solutionTimestamp = timestamp
		forging.solutionAddress = address
	}

}

// thread not safe
func (forging *Forging) publishSolution() (err error) {

	forging.blkComplete.Block.Forger = forging.solutionAddress.publicKeyHash
	forging.blkComplete.Block.DelegatedPublicKey = forging.solutionAddress.delegatedPublicKey
	forging.blkComplete.Block.Timestamp = forging.solutionTimestamp
	if forging.blkComplete.Block.Height > 0 {
		forging.blkComplete.Block.StakingAmount = forging.solutionAddress.account.GetDelegatedStakeAvailable(forging.blkComplete.Block.Height)
	}

	serializationForSigning := forging.blkComplete.Block.SerializeForSigning()

	if forging.blkComplete.Block.Signature, err = forging.solutionAddress.delegatedPrivateKey.Sign(&serializationForSigning); err != nil {
		gui.Error("Error signing forged block", err)
		return
	}

	//send message to blockchain
	forging.SolutionChannel <- forging.blkComplete
	return
}
