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

func (self *KnownNodeScored) IncrementScore(update int32, isServer bool) bool {

	newScore := atomic.AddInt32(&self.Score, update)
	if newScore < -100 && !self.IsSeed {
		return true
	}
	if newScore > 100 && !isServer {
		atomic.StoreInt32(&self.Score, 100)
	}
	if newScore > 300 && isServer {
		atomic.StoreInt32(&self.Score, 300)
	}

	return false
}
