package forging

import (
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/block"
	"pandora-pay/config"
	"pandora-pay/gui"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var started int32
var forgingWorking int32

type forgingType struct {
	blkComplete *block.BlockComplete

	solution          bool
	solutionTimestamp uint64
	solutionAddress   *ForgingWalletAddress
}

var forging forgingType

func ForgingInit() {

	gui.Log("Forging Init")
	if err := ForgingW.loadBalances(); err != nil {
		gui.Error("Error reading balances", err)
	}
	go startForging(config.CPU_THREADS)

}

func startForging(threads int) {

	var err error

	if !atomic.CompareAndSwapInt32(&started, 0, 1) {
		return
	}

	for atomic.LoadInt32(&started) == 1 {

		if forging.blkComplete == nil {

			forging.blkComplete, err = createNextBlockComplete(blockchain.Chain.Height)
			if err != nil {
				gui.Error("Error creating new block", err)
				time.Sleep(5 * time.Second)
			}

		}
		if !atomic.CompareAndSwapInt32(&forgingWorking, 0, 1) {
			gui.Error("A strange error as forgingWorking couldn't be set to 1 ")
			return
		}

		wg := sync.WaitGroup{}

		for i := 0; i < threads; i++ {
			wg.Add(1)
			go forge(threads, i, &wg)
		}

		wg.Wait()

		if forging.solution && forging.blkComplete != nil {
			forging.publishSolution()
			forging.blkComplete = nil
		}

	}

}

func stopForging() {
	atomic.AddInt32(&started, -1)
}

//thread safe
func RestartForging(stakeNewBlock bool) {

	atomic.CompareAndSwapInt32(&forgingWorking, 1, 0)

	if stakeNewBlock == true {
		forging.blkComplete = nil
	}

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

	forging.blkComplete.Block.Forger = forging.solutionAddress.delegatedPublicKey
	forging.blkComplete.Block.Timestamp = forging.solutionTimestamp
	serializationForSigning := forging.blkComplete.Block.SerializeForSigning()

	signature, _ := forging.solutionAddress.delegatedPrivateKey.Sign(&serializationForSigning)

	copy(forging.blkComplete.Block.Signature[:], signature)

	var array []*block.BlockComplete
	array = append(array, forging.blkComplete)

	result, err := blockchain.Chain.AddBlocks(array)
	if err == nil && result {
		gui.Info("Block was forged! " + strconv.FormatUint(forging.blkComplete.Block.Height, 10))
	} else {
		gui.Error("Error forging block "+strconv.FormatUint(forging.blkComplete.Block.Height, 10), err)
	}

}
