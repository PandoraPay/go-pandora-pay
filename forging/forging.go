package forging

import (
	"pandora-pay/blockchain"
	"pandora-pay/config"
	"pandora-pay/gui"
	"sync"
	"time"
)

var started = false
var forging bool

func ForgingInit() {

	gui.Log("Forging Init")
	startForging(config.CPU_THREADS)

}

func startForging(threads int) {

	started = true
	for started {

		blkComplete, err := createNextBlockComplete(blockchain.Chain.Height)
		if err != nil {
			gui.Error("Error creating new block", err)
			time.Sleep(5 * time.Second)
		}

		wg := sync.WaitGroup{}
		wg.Add(threads)

		forging = true
		for i := 0; i < threads; i++ {
			go forge(blkComplete, threads, i, &wg)
		}

		wg.Wait()

	}

}

func stopForging() {
	started = false
}
