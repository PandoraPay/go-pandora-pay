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
	knownMap  map[string]*KnownNode
	knownList []*KnownNode
	sync.RWMutex
}

func (self *KnownNodes) GetRandomKnownNode() *KnownNode {
	self.RLock()
	defer self.RUnlock()

	return self.knownList[rand.Intn(len(self.knownList))]
}

func (self *KnownNodes) AddKnownNode(url *url.URL, isSeed bool) bool {
	self.Lock()
	defer self.Unlock()

	urlString := url.String()

	if self.knownMap[urlString] != nil {
		return false
	}

	knownNode := &KnownNode{
		Url:         url,
		UrlStr:      urlString,
		UrlHostOnly: url.Host,
		IsSeed:      isSeed,
	}
	self.knownList = append(self.knownList, knownNode)
	self.knownMap[urlString] = knownNode

	return true
}

func (self *KnownNodes) RemoveKnownNode(knownNode *KnownNode) {
	self.Lock()
	defer self.Unlock()

	if self.knownMap[knownNode.UrlStr] != nil {
		delete(self.knownMap, knownNode.UrlStr)
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
		make(map[string]*KnownNode),
		make([]*KnownNode, 0),
		sync.RWMutex{},
	}

	return
}
