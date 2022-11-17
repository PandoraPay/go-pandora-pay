package consensus

import (
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
)

type Forks struct {
	hashes *generics.Map[string, *Fork]
}

func (forks *Forks) getBestFork() (selectedFork *Fork) {

	bigTotalDifficulty := config.BIG_INT_ZERO

	forks.hashes.Range(func(key string, fork *Fork) bool {

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

func (forks *Forks) addFork(fork *Fork) {
	forks.hashes.LoadOrStore(fork.HashStr, fork)
}

func (forks *Forks) removeFork(fork *Fork) {
	forks.hashes.Delete(fork.HashStr)
}
