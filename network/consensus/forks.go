package consensus

import (
	"pandora-pay/config"
	"sync"
)

type Forks struct {
	hashes       sync.Map
	list         []*Fork
	sync.RWMutex `json:"-"`
}

func (forks *Forks) getBestFork() (selectedFork *Fork) {
	forks.RLock()
	defer forks.RUnlock()
	if len(forks.list) > 0 {
		bigTotalDifficulty := config.BIG_INT_ZERO
		for _, fork := range forks.list {
			fork.RLock()
			if !fork.ready && fork.bigTotalDifficulty.Cmp(bigTotalDifficulty) > 0 {
				bigTotalDifficulty = fork.bigTotalDifficulty
				selectedFork = fork
			}
			fork.RUnlock()
		}
	}
	return
}

func (forks *Forks) removeFork(fork *Fork) {
	forks.RLock()
	defer forks.RUnlock()
	for i, fork2 := range forks.list {
		if fork == fork2 {
			//order is not important
			forks.list[i] = forks.list[len(forks.list)-1]
			forks.list = forks.list[:len(forks.list)-1]
		}
	}

}
