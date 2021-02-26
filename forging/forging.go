package forging

import (
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/wallet"
	"sync"
	"time"
)

var started = false

type forgingType struct {
	processing bool

	solutionTimestamp uint64
	solutionPublicKey [33]byte

	sync.RWMutex
}

var forging forgingType

func ForgingInit() {

	gui.Log("Forging Init")
	go startForging(config.CPU_THREADS)

}

func startForging(threads int) {

	started = true
	for started {

		blkComplete, err := createNextBlockComplete(blockchain.Chain.Height)
		if err != nil {
			gui.Error("Error creating new block", err)
			time.Sleep(5 * time.Second)
		}

		wallet.W.RLock()
		addresses := wallet.W.Addresses

		forging.Lock()
		forging.processing = true
		forging.Unlock()

		wg := sync.WaitGroup{}
		wg.Add(threads)

		for i := 0; i < threads; i++ {
			go forge(blkComplete, threads, i, &wg, addresses)
		}

		wg.Wait()
		wallet.W.RUnlock()

	}

}

func stopForging() {
	started = false
}

func foundSolution(publicKey [33]byte, timestamp uint64) {

	forging.Lock()
	defer forging.Unlock()

	if forging.processing {
		return
	}

	forging.processing = false
	forging.solutionTimestamp = timestamp

	//copy publicKey
	//to make it thread safe and unlock the wallet
	copy(forging.solutionPublicKey[:], publicKey[:])

}

func (forging *forgingType) safeIsProcessing() (r bool) {
	forging.RLock()
	r = forging.processing
	forging.RUnlock()
	return
}
