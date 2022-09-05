package known_node

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
		return false
	}
	if newScore > 300 && isServer {
		atomic.StoreInt32(&self.Score, 300)
		return false
	}

	return true
}

func (self *KnownNodeScored) DecreaseScore(delta int32, isServer bool) (bool, bool) {

	newScore := atomic.AddInt32(&self.Score, delta)
	if newScore < -100 {
		if !self.IsSeed {
			return true, true
		}
		atomic.StoreInt32(&self.Score, -100)
		return false, false
	}
	return true, false
}
