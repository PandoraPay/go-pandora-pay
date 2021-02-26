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

type forgingType struct {
	BlkComplete *block.BlockComplete
	target      *big.Int

	solution          bool
	solutionTimestamp uint64
	solutionAddress   *ForgingWalletAddress

	SolutionChannel chan int
}

var Forging forgingType
var wg = sync.WaitGroup{}

func ForgingInit() {

	gui.Log("Forging Init")
	if err := ForgingW.loadBalances(); err != nil {
		gui.Error("Error reading balances", err)
	}

	Forging.SolutionChannel = make(chan int)

	go startForging(config.CPU_THREADS)

}

func startForging(threads int) {

	if !atomic.CompareAndSwapInt32(&started, 0, 1) {
		return
	}

	for atomic.LoadInt32(&started) == 1 {

		if Forging.solution || Forging.BlkComplete == nil {
			// gui.Error("No block for staking..." )
			time.Sleep(10 * time.Millisecond)
		}

		if !atomic.CompareAndSwapInt32(&forgingWorking, 0, 1) {
			gui.Error("A strange error as forgingWorking couldn't be set to 1 ")
			return
		}

		for i := 0; i < threads; i++ {
			wg.Add(1)
			go forge(threads, i)
		}

		wg.Wait()

		if Forging.solution && Forging.BlkComplete != nil {
			Forging.publishSolution()
		}

	}

}

func stopForging() {
	atomic.AddInt32(&started, -1)
}

func StopForgingWorkers() {
	atomic.CompareAndSwapInt32(&forgingWorking, 1, 0)
}

//thread safe
func (forging *forgingType) RestartForgingWorkers(BlkComplete *block.BlockComplete, target *big.Int) {

	atomic.CompareAndSwapInt32(&forgingWorking, 1, 0)

	wg.Wait()

	forging.solution = false
	forging.BlkComplete = BlkComplete
	forging.target = target

}

//thread safe
func (forging *forgingType) foundSolution(address *ForgingWalletAddress, timestamp uint64) {

	if atomic.CompareAndSwapInt32(&forgingWorking, 1, 0) {
		forging.solution = true
		forging.solutionTimestamp = timestamp
		forging.solutionAddress = address
	}

}

// thread not safe
func (forging *forgingType) publishSolution() {

	forging.BlkComplete.Block.Forger = forging.solutionAddress.delegatedPublicKey
	forging.BlkComplete.Block.Timestamp = forging.solutionTimestamp
	serializationForSigning := forging.BlkComplete.Block.SerializeForSigning()

	signature, _ := forging.solutionAddress.delegatedPrivateKey.Sign(&serializationForSigning)

	copy(forging.BlkComplete.Block.Signature[:], signature)

	//send message to blockchain
	Forging.SolutionChannel <- 1
}
