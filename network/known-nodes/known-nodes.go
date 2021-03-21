package known_nodes

import (
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
	KnownMap     sync.Map
	KnownList    []*KnownNode
	sync.RWMutex `json:"-"`
}

func (knownNodes *KnownNodes) AddKnownNode(url *url.URL, isSeed bool) (result bool, knownNode *KnownNode) {

	urlString := url.String()
	var exists bool
	var found interface{}
	if found, exists = knownNodes.KnownMap.Load(urlString); exists {
		knownNode = found.(*KnownNode)
		return false, knownNode
	}

	knownNodes.Lock()
	defer knownNodes.Unlock()

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
	knownNodes.KnownList = append(knownNodes.KnownList, knownNode)

	result = true
	return
}

func CreateKnownNodes() *KnownNodes {

	return &KnownNodes{
		KnownMap:  sync.Map{},
		KnownList: []*KnownNode{},
	}

}
