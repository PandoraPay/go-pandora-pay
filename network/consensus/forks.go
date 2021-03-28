package consensus

import (
	"pandora-pay/config"
	"sync"
)

type Forks struct {
	hashes           *sync.Map
	forksDownloadMap *sync.Map //*Fork
	id               uint32
}

func (forks *Forks) getBestFork(forksMap *sync.Map) (selectedFork *Fork) {

	bigTotalDifficulty := config.BIG_INT_ZERO
	forksMap.Range(func(key, value interface{}) bool {
		fork := value.(*Fork)
		if !fork.readyForDownloading.IsSet() && fork.bigTotalDifficulty.Cmp(bigTotalDifficulty) > 0 {
			bigTotalDifficulty = fork.bigTotalDifficulty
			selectedFork = fork
		}
		return true
	})
	return
}

func (forks *Forks) removeFork(fork *Fork) {

	for _, hash := range fork.hashes {
		forks.hashes.Delete(string(hash))
	}

}
