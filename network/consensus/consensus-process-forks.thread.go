package consensus

import (
	"bytes"
	api_store "pandora-pay/network/api/api-store"
	api_websockets "pandora-pay/network/api/api-websockets"
	"time"
)

type ConsensusProcessForksThread struct {
	forks    *Forks
	apiStore *api_store.APIStore
}

func (thread *ConsensusProcessForksThread) processFork(fork *Fork) {

	fork.Lock()
	defer fork.Lock()

	if !fork.ready {
		return
	}
	prevHash := fork.prevHash

	for i := fork.start; i >= 0; i-- {

		fork2Data, exists := thread.forks.hashes.LoadOrStore(string(prevHash), fork)
		if exists { //let's merge
			fork2 := fork2Data.(*Fork)
			if fork2.mergeFork(fork) {
				thread.forks.removeFork(fork)
				return
			}
		}

		conn := fork.conns[0]
		answer := conn.SendAwaitAnswer([]byte("block-complete"), api_websockets.APIBlockHeight(i-1))
		if answer.Err != nil {
			fork.errors += 1
			if fork.errors > 2 {
				thread.forks.removeFork(fork)
				return
			}
		} else {
			prevHash := answer.Out

			chainHash, err := thread.apiStore.LoadBlockHash(i - 1)
			if err == nil {
				if bytes.Equal(prevHash, chainHash) {
					fork.ready = true
					return
				}
			}

			fork.start -= 1
			if fork.errors >= -10 {
				fork.errors -= 1
			}
		}
	}

}

func (thread *ConsensusProcessForksThread) execute() {
	for {

		fork := thread.forks.getBestFork()
		if fork != nil {
			thread.processFork(fork)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func createConsensusProcessForksThread(forks *Forks, apiStore *api_store.APIStore) *ConsensusProcessForksThread {
	return &ConsensusProcessForksThread{
		forks:    forks,
		apiStore: apiStore,
	}
}
