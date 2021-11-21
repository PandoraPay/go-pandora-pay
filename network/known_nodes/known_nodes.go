package known_nodes

import (
	"math/rand"
	"sync"
)

type KnownNode struct {
	UrlStr string
	IsSeed bool
}

type KnownNodeScored struct {
	KnownNode
	Score int
	sync.RWMutex
}

type KnownNodes struct {
	knownMap       *sync.Map //*KnownNode
	knownList      []*KnownNodeScored
	knownListMutex sync.RWMutex
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

func (self *KnownNodes) AddKnownNode(url string, isSeed bool) bool {

	knownNode := &KnownNodeScored{
		KnownNode: KnownNode{
			UrlStr: url,
			IsSeed: isSeed,
		},
		Score: 0,
	}

	if _, exists := self.knownMap.LoadOrStore(url, knownNode); exists {
		return false
	}

	self.knownListMutex.Lock()
	self.knownList = append(self.knownList, knownNode)
	self.knownListMutex.Unlock()

	return true
}

func (self *KnownNodes) RemoveKnownNode(knownNode *KnownNodeScored) {

	if _, exists := self.knownMap.LoadAndDelete(knownNode.UrlStr); exists {
		self.knownListMutex.Lock()
		defer self.knownListMutex.Unlock()
		for i, knownNode2 := range self.knownList {
			if knownNode2 == knownNode {
				self.knownList[i] = self.knownList[len(self.knownList)-1]
				self.knownList = self.knownList[:len(self.knownList)-1]
				return
			}
		}
	}
}

func CreateKnownNodes() (knownNodes *KnownNodes) {

	knownNodes = &KnownNodes{
		&sync.Map{},
		make([]*KnownNodeScored, 0),
		sync.RWMutex{},
	}

	return
}
