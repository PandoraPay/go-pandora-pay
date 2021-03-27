package consensus

import (
	"bytes"
	block_complete "pandora-pay/blockchain/block-complete"
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
	defer fork.Unlock()

	if fork.ready {
		return
	}

	var err error
	prevHash := fork.prevHash

	for fork.start >= 0 {

		if fork.errors > 2 {
			thread.forks.removeFork(fork)
			return
		}
		if fork.errors >= -10 {
			fork.errors -= 1
		}

		fork2Data, exists := thread.forks.hashes.LoadOrStore(string(prevHash), fork)
		if exists { //let's merge
			fork2 := fork2Data.(*Fork)
			if fork2.mergeFork(fork) {
				thread.forks.removeFork(fork)
				return
			}
		}

		conn := fork.conns[0]
		answer := conn.SendAwaitAnswer([]byte("block-complete"), api_websockets.APIBlockHeight(fork.start-1))
		if answer.Err != nil {
			fork.errors += 1
			continue
		}

		blkComplete := block_complete.CreateEmptyBlockComplete()
		if err = blkComplete.Deserialize(answer.Out); err != nil {
			fork.errors += 1
			continue
		}
		if err = blkComplete.BloomAll(true, true, true); err != nil {
			fork.errors += 1
			continue
		}

		chainHash, err := thread.apiStore.LoadBlockHash(fork.start - 1)
		if err == nil {
			if bytes.Equal(prevHash, chainHash) {
				fork.ready = true
				return
			}
		}

		fork.start -= 1
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
