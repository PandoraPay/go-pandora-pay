package known_nodes

import (
	"errors"
	"math/rand"
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
	"sync"
	"sync/atomic"
)

type KnownNodes struct {
	knownMap       *generics.Map[string, *KnownNodeScored]
	knownList      []*KnownNodeScored
	knownListMutex sync.RWMutex
	knownCount     int32 //atomic required
}

func (self *KnownNodes) GetList() []*KnownNodeScored {
	self.knownListMutex.RLock()
	knownList := make([]*KnownNodeScored, len(self.knownList))
	for i, knowNode := range self.knownList {
		knownList[i] = knowNode
	}
	self.knownListMutex.RUnlock()

	return knownList
}

func (self *KnownNodes) GetRandomKnownNode() *KnownNodeScored {
	self.knownListMutex.RLock()
	defer self.knownListMutex.RUnlock()
	return self.knownList[rand.Intn(len(self.knownList))]
}

func (self *KnownNodes) AddKnownNode(url string, isSeed bool) (*KnownNodeScored, error) {

	if atomic.LoadInt32(&self.knownCount) > config.NETWORK_KNOWN_NODES_LIMIT {
		return nil, errors.New("Too many nodes already in the list")
	}

	knownNode := &KnownNodeScored{
		KnownNode: KnownNode{
			URL:    url,
			IsSeed: isSeed,
		},
		Score: 0,
	}

	if _, exists := self.knownMap.LoadOrStore(url, knownNode); exists {
		return nil, errors.New("Already exists")
	}

	self.knownListMutex.Lock()
	self.knownList = append(self.knownList, knownNode)
	self.knownListMutex.Unlock()

	atomic.AddInt32(&self.knownCount, +1)

	return knownNode, nil
}

func (self *KnownNodes) RemoveKnownNode(knownNode *KnownNodeScored) {

	if _, exists := self.knownMap.LoadAndDelete(knownNode.URL); exists {
		self.knownListMutex.Lock()
		defer self.knownListMutex.Unlock()
		for i, knownNode2 := range self.knownList {
			if knownNode2 == knownNode {
				self.knownList[i] = self.knownList[len(self.knownList)-1]
				self.knownList = self.knownList[:len(self.knownList)-1]
				atomic.AddInt32(&self.knownCount, -1)
				return
			}
		}
	}

}

func NewKnownNodes() (knownNodes *KnownNodes) {

	knownNodes = &KnownNodes{
		&generics.Map[string, *KnownNodeScored]{},
		make([]*KnownNodeScored, 0),
		sync.RWMutex{},
		0,
	}

	return
}
