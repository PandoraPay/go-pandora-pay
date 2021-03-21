package consensus

import "sync"

type Forks struct {
	hashes       sync.Map
	list         []*Fork
	sync.RWMutex `json:"-"`
}
