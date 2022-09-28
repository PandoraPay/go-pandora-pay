package known_nodes

import (
	"errors"
	"math/rand"
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes/known_node"
	"pandora-pay/store/min_max_heap"
	"sync"
	"sync/atomic"
)

type KnownNodes struct {
	connectedNodes                *connected_nodes.ConnectedNodes
	bannedNodes                   *banned_nodes.BannedNodes
	knownMap                      *generics.Map[string, *known_node.KnownNodeScored]
	knownList                     []*known_node.KnownNodeScored //contains all known peers
	knownListMutex                sync.RWMutex
	knownNotConnectedMaxHeap      *min_max_heap.HeapMemory //contains known peers that we are not connected
	knownNotConnectedMaxHeapMutex sync.RWMutex
	knownCount                    int32 //atomic required
}

func (self *KnownNodes) GetList() []*known_node.KnownNodeScored {
	self.knownListMutex.RLock()
	defer self.knownListMutex.RUnlock()

	knownList := make([]*known_node.KnownNodeScored, len(self.knownList))
	for i, knowNode := range self.knownList {
		knownList[i] = knowNode
	}

	return knownList
}

func (self *KnownNodes) GetRandomKnownNode() *known_node.KnownNodeScored {
	self.knownListMutex.RLock()
	defer self.knownListMutex.RUnlock()
	if len(self.knownList) == 0 {
		return nil
	}
	return self.knownList[rand.Intn(len(self.knownList))]
}

func (self *KnownNodes) GetBestNotConnectedKnownNode() *known_node.KnownNodeScored {
	self.knownNotConnectedMaxHeapMutex.RLock()
	top, _ := self.knownNotConnectedMaxHeap.GetTop()
	self.knownNotConnectedMaxHeapMutex.RUnlock()
	if top == nil {
		return nil
	}
	if top.Score == float64(known_node.KNOWN_KNODE_SCORE_MINIMUM) {
		return self.GetRandomKnownNode()
	}
	knownNode, _ := self.knownMap.Load(string(top.Key))
	return knownNode
}

func (self *KnownNodes) IncreaseKnownNodeScore(knownNode *known_node.KnownNodeScored, delta int32, isServer bool) bool {
	update, score := knownNode.IncreaseScore(delta, isServer)
	if update {
		self.knownNotConnectedMaxHeapMutex.Lock()
		defer self.knownNotConnectedMaxHeapMutex.Unlock()
		self.knownNotConnectedMaxHeap.Update(float64(score), []byte(knownNode.URL))
	}
	return update
}

func (self *KnownNodes) DecreaseKnownNodeScore(knownNode *known_node.KnownNodeScored, delta int32, isServer bool) (bool, bool) {
	update, removed, score := knownNode.DecreaseScore(delta, isServer)
	if removed {
		self.RemoveKnownNode(knownNode)
	}
	if update || removed {
		self.knownNotConnectedMaxHeapMutex.Lock()
		defer self.knownNotConnectedMaxHeapMutex.Unlock()
		self.knownNotConnectedMaxHeap.DeleteByKey([]byte(knownNode.URL))
		if !removed {
			self.knownNotConnectedMaxHeap.Insert(float64(score), []byte(knownNode.URL))
		}
	}
	return update, removed
}

func (self *KnownNodes) MarkKnownNodeConnected(knownNode *known_node.KnownNodeScored) {
	self.knownNotConnectedMaxHeapMutex.Lock()
	defer self.knownNotConnectedMaxHeapMutex.Unlock()
	self.knownNotConnectedMaxHeap.DeleteByKey([]byte(knownNode.URL))
}

func (self *KnownNodes) MarkKnownNodeDisconnected(knownNode *known_node.KnownNodeScored) {
	self.knownNotConnectedMaxHeapMutex.Lock()
	defer self.knownNotConnectedMaxHeapMutex.Unlock()
	self.knownNotConnectedMaxHeap.Update(float64(atomic.LoadInt32(&knownNode.Score)), []byte(knownNode.URL))
}

func (self *KnownNodes) AddKnownNode(url string, isSeed bool) (*known_node.KnownNodeScored, error) {

	if url == "" {
		return nil, errors.New("url is empty")
	}

	if atomic.LoadInt32(&self.knownCount) > config.NETWORK_KNOWN_NODES_LIMIT {
		return nil, errors.New("Too many nodes already in the list")
	}

	if self.bannedNodes.IsBanned(url) {
		return nil, errors.New("url is banned")
	}

	knownNode := &known_node.KnownNodeScored{
		KnownNode: known_node.KnownNode{
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

	if _, ok := self.connectedNodes.AllAddresses.Load(url); !ok {
		self.knownNotConnectedMaxHeapMutex.Lock()
		self.knownNotConnectedMaxHeap.Update(0, []byte(url))
		self.knownNotConnectedMaxHeapMutex.Unlock()
	}

	return knownNode, nil
}

func (self *KnownNodes) RemoveKnownNode(knownNode *known_node.KnownNodeScored) {

	if _, exists := self.knownMap.LoadAndDelete(knownNode.URL); exists {

		self.knownNotConnectedMaxHeapMutex.Lock()
		self.knownNotConnectedMaxHeap.DeleteByKey([]byte(knownNode.URL))
		self.knownNotConnectedMaxHeapMutex.Unlock()

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

func NewKnownNodes(connectedNodes *connected_nodes.ConnectedNodes, bannedNodes *banned_nodes.BannedNodes) (knownNodes *KnownNodes) {

	knownNodes = &KnownNodes{
		connectedNodes,
		bannedNodes,
		&generics.Map[string, *known_node.KnownNodeScored]{},
		make([]*known_node.KnownNodeScored, 0),
		sync.RWMutex{},
		min_max_heap.NewMaxMemoryHeap(),
		sync.RWMutex{},
		0,
	}

	return
}
