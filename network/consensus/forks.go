package consensus

import (
	"pandora-pay/config"
	"sync"
)

type Forks struct {
	hashes *sync.Map
}

func (forks *Forks) getBestFork() (selectedFork *Fork) {

	bigTotalDifficulty := config.BIG_INT_ZERO

	forks.hashes.Range(func(key interface{}, value interface{}) bool {

		fork := value.(*Fork)
		fork.RLock()
		if fork.BigTotalDifficulty.Cmp(bigTotalDifficulty) > 0 {
			bigTotalDifficulty = fork.BigTotalDifficulty
			selectedFork = fork
		}
		fork.RUnlock()

		return true
	})

	return
}

func (forks *Forks) removeFork(fork *Fork) {
	forks.hashes.Delete(fork.HashStr)
}
