package known_nodes

import (
	"math/rand"
	"net/url"
	"sync"
)

type KnownNode struct {
	Url         *url.URL
	UrlStr      string
	UrlHostOnly string
	IsSeed      bool
}

type KnownNodes struct {
	knownMap       *sync.Map //*KnownNode
	knownList      []*KnownNode
	knownListMutex sync.RWMutex
}

func (self *KnownNodes) GetRandomKnownNode() *KnownNode {
	self.knownListMutex.RLock()
	defer self.knownListMutex.RUnlock()
	return self.knownList[rand.Intn(len(self.knownList))]
}

func (self *KnownNodes) AddKnownNode(url *url.URL, isSeed bool) bool {

	urlString := url.String()
	knownNode := &KnownNode{
		Url:         url,
		UrlStr:      urlString,
		UrlHostOnly: url.Host,
		IsSeed:      isSeed,
	}

	if _, exists := self.knownMap.LoadOrStore(urlString, knownNode); exists {
		return false
	}

	self.knownListMutex.Lock()
	defer self.knownListMutex.Unlock()
	self.knownList = append(self.knownList, knownNode)

	return true
}

func (self *KnownNodes) RemoveKnownNode(knownNode *KnownNode) {

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
		make([]*KnownNode, 0),
		sync.RWMutex{},
	}

	return
}
