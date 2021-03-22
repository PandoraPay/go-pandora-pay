package consensus

import (
	"math/rand"
	"sync"
)

type Forks struct {
	hashes       sync.Map
	list         []*Fork
	sync.RWMutex `json:"-"`
}

func (forks *Forks) getRandomFork() *Fork {
	forks.RLock()
	defer forks.RUnlock()
	if len(forks.list) > 0 {
		return forks.list[rand.Intn(len(forks.list))]
	}
	return nil
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
