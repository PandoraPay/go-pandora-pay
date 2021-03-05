package forging

import (
	"math/big"
	"pandora-pay/blockchain/block"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sync"
	"sync/atomic"
	"time"
)

type Forging struct {
	blkComplete *block.BlockComplete
	target      *big.Int

	solution          bool
	solutionTimestamp uint64
	solutionAddress   *ForgingWalletAddress

	started        int32
	forgingWorking int32

	wg sync.WaitGroup

	Wallet          ForgingWallet
	SolutionChannel chan *block.BlockComplete
}

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

	if !atomic.CompareAndSwapInt32(&forging.started, 0, 1) {
		return
	}

	for atomic.LoadInt32(&forging.started) == 1 {

		if forging.solution || forging.blkComplete == nil {
			// gui.Error("No block for staking..." )
			time.Sleep(10 * time.Millisecond)
			continue
		}

		if !atomic.CompareAndSwapInt32(&forging.forgingWorking, 0, 1) {
			gui.Error("A strange error as forgingWorking couldn't be set to 1 ")
			return
		}

		for i := 0; i < threads; i++ {
			forging.wg.Add(1)
			go forge(forging, threads, i)
		}

		forging.wg.Wait()

		if forging.solution {
			err := forging.publishSolution()
			if err != nil {
				forging.solution = false
			}
		}

	}

}

func (forging *Forging) StopForging() {
	forging.StopForgingWorkers()
	atomic.AddInt32(&forging.started, -1)
}

func (forging *Forging) StopForgingWorkers() {
	atomic.CompareAndSwapInt32(&forging.forgingWorking, 1, 0)
}

//thread safe
func (forging *Forging) RestartForgingWorkers(blkComplete *block.BlockComplete, target *big.Int) {

	atomic.CompareAndSwapInt32(&forging.forgingWorking, 1, 0)

	forging.wg.Wait()

	forging.solution = false
	forging.blkComplete = blkComplete
	forging.target = target

}

//thread safe
func (forging *Forging) foundSolution(address *ForgingWalletAddress, timestamp uint64) {

	if atomic.CompareAndSwapInt32(&forging.forgingWorking, 1, 0) {
		forging.solution = true
		forging.solutionTimestamp = timestamp
		forging.solutionAddress = address
	}

}

// thread not safe
func (forging *Forging) publishSolution() (err error) {

	defer func() {
		if err2 := recover(); err2 != nil {
			err = helpers.ConvertRecoverError(err2)
			gui.Error("Error signing forged block", err)
		}
	}()

	forging.blkComplete.Block.Forger = forging.solutionAddress.publicKeyHash
	forging.blkComplete.Block.DelegatedPublicKey = forging.solutionAddress.delegatedPublicKey
	forging.blkComplete.Block.Timestamp = forging.solutionTimestamp
	if forging.blkComplete.Block.Height > 0 {
		forging.blkComplete.Block.StakingAmount = forging.solutionAddress.account.GetDelegatedStakeAvailable(forging.blkComplete.Block.Height)
	}

	serializationForSigning := forging.blkComplete.Block.SerializeForSigning()

	forging.blkComplete.Block.Signature = forging.solutionAddress.delegatedPrivateKey.Sign(serializationForSigning)

	//send message to blockchain
	forging.SolutionChannel <- forging.blkComplete
	return
}

func (forging *Forging) Close() {
	forging.StopForging()
}
