package known_nodes

import (
	"net/url"
	"sync"
	"sync/atomic"
)

type KnownNode struct {
	Url         *url.URL
	UrlStr      string
	UrlHostOnly string
	IsSeed      bool
}

type KnownNodes struct {
	KnownMap       sync.Map
	KnownList      *atomic.Value //[]*KnownNode
	KnownListMutex *sync.Mutex
}

func (knownNodes *KnownNodes) AddKnownNode(url *url.URL, isSeed bool) (result bool, knownNode *KnownNode) {

	urlString := url.String()
	var exists bool
	var found interface{}
	if found, exists = knownNodes.KnownMap.Load(urlString); exists {
		knownNode = found.(*KnownNode)
		return false, knownNode
	}

	knownNode = &KnownNode{
		Url:         url,
		UrlStr:      urlString,
		UrlHostOnly: url.Host,
		IsSeed:      isSeed,
	}

	if found, exists := knownNodes.KnownMap.LoadOrStore(urlString, knownNode); exists {
		knownNode = found.(*KnownNode)
		return false, knownNode
	}

	knownNodes.KnownListMutex.Lock()
	knownList := knownNodes.KnownList.Load().([]*KnownNode)
	knownNodes.KnownList.Store(append(knownList, knownNode))
	knownNodes.KnownListMutex.Unlock()

	result = true
	return
}

func CreateKnownNodes() (knownNodes *KnownNodes) {

	knownNodes = &KnownNodes{
		KnownMap:       sync.Map{},
		KnownList:      &atomic.Value{},
		KnownListMutex: &sync.Mutex{},
	}

	knownNodes.KnownList.Store([]*KnownNode{})

	return
}
