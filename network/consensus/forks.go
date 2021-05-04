package consensus

import (
	"math/big"
	"pandora-pay/config"
	"sync"
	"sync/atomic"
)

type Forks struct {
	hashes    *sync.Map
	list      *atomic.Value //[]*Fork
	listMutex *sync.Mutex
}

func (forks *Forks) getBestFork() (selectedFork *Fork) {

	bigTotalDifficulty := config.BIG_INT_ZERO
	list := forks.list.Load().([]*Fork)

	for _, fork := range list {

		forkBigTotalDifficulty := fork.bigTotalDifficulty.Load().(*big.Int)
		if forkBigTotalDifficulty.Cmp(bigTotalDifficulty) > 0 {
			bigTotalDifficulty = forkBigTotalDifficulty
			selectedFork = fork
		}
	}

	return
}

func (forks *Forks) removeFork(fork *Fork, removeHashes bool) {

	fork.RLock()
	defer fork.RUnlock()

	forks.listMutex.Lock()
	defer forks.listMutex.Unlock()

	list := forks.list.Load().([]*Fork)
	for i, fork2 := range list {
		if fork2 == fork {
			list[i] = list[len(list)-1]
			list = list[:len(list)-1]

			forks.list.Store(list)
			if removeHashes {
				forks.hashes.Delete(string(fork.hash))
			}

			break
		}
	}

}
