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

var KNOWN_KNODE_SCORE_MINIMUM = int32(-1000)

func (self *KnownNodeScored) IncreaseScore(delta int32, isServer bool) (bool, int32) {

	newScore := atomic.AddInt32(&self.Score, delta)

	if newScore > 100 && !isServer {
		atomic.StoreInt32(&self.Score, 100)
		return false, 100
	}
	if newScore > 300 && isServer {
		atomic.StoreInt32(&self.Score, 300)
		return false, 100
	}

	return true, newScore
}

func (self *KnownNodeScored) DecreaseScore(delta int32, isServer bool) (bool, bool, int32) {

	newScore := atomic.AddInt32(&self.Score, delta)
	if newScore < KNOWN_KNODE_SCORE_MINIMUM {
		if !self.IsSeed {
			return true, true, KNOWN_KNODE_SCORE_MINIMUM
		}
		atomic.StoreInt32(&self.Score, KNOWN_KNODE_SCORE_MINIMUM)
		return true, false, KNOWN_KNODE_SCORE_MINIMUM
	}
	return true, false, newScore
}
