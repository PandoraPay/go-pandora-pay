package forging

import (
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/gui"
	"sync"
	"time"
)

var started = false

func ForgingInit() {

	gui.Log("Forging Init")
	startForging(config.CPU_THREADS)

}

func startForging(threads int) {

	started = true
	for started {

		block, err := createNextBlock(blockchain.Chain.Height)
		if err != nil {
			gui.Error("Error creating new block", err)
			time.Sleep(5 * time.Second)
		}

		wg := sync.WaitGroup{}
		wg.Add(threads)

		for i := 0; i < threads; i++ {
			go forge(block, threads, i, &wg)
		}

		wg.Wait()

		time.Sleep(1 * time.Second)
	}

}

func stopForging() {
	started = false
}
