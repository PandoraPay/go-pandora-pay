package known_nodes

import (
	"sync/atomic"
)

type KnownNode struct {
	URL    string
	IsSeed bool
}

type KnownNodeScored struct {
	KnownNode
	Score int32 //use atomic
}

func (self *KnownNodeScored) IncreaseScore(delta int32, isServer bool) bool {

	newScore := atomic.AddInt32(&self.Score, delta)

	if newScore > 100 && !isServer {
		atomic.StoreInt32(&self.Score, 100)
		return true
	}
	if newScore > 300 && isServer {
		atomic.StoreInt32(&self.Score, 300)
		return true
	}

	return false
}

func (self *KnownNodeScored) DecreaseScore(delta int32, isServer bool) bool {

	newScore := atomic.AddInt32(&self.Score, delta)
	if newScore < -100 {
		if !self.IsSeed {
			return true
		}
		atomic.StoreInt32(&self.Score, -100)
	}
	return false
}
