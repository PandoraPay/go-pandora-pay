package known_nodes

import (
	"errors"
	"math/rand"
	"pandora-pay/helpers/generics"
	"pandora-pay/network/banned_nodes"
	"pandora-pay/network/connected_nodes"
	"pandora-pay/network/known_nodes/known_node"
	"pandora-pay/network/network_config"
	"pandora-pay/store/min_max_heap"
	"sync"
	"sync/atomic"
)

type KnownNodesType struct {
	knownMap                      *generics.Map[string, *known_node.KnownNodeScored]
	knownList                     []*known_node.KnownNodeScored //contains all known peers
	knownListMutex                sync.RWMutex
	knownNotConnectedMaxHeap      *min_max_heap.HeapMemory //contains known peers that we are not connected
	knownNotConnectedMaxHeapMutex sync.RWMutex
	knownCount                    int32 //atomic required
}

var KnownNodes *KnownNodesType

func (this *KnownNodesType) GetList() []*known_node.KnownNodeScored {
	this.knownListMutex.RLock()
	defer this.knownListMutex.RUnlock()

	knownList := make([]*known_node.KnownNodeScored, len(this.knownList))
	for i, knowNode := range this.knownList {
		knownList[i] = knowNode
	}

	return knownList
}

func (this *KnownNodesType) GetRandomKnownNode() *known_node.KnownNodeScored {
	this.knownListMutex.RLock()
	defer this.knownListMutex.RUnlock()
	if len(this.knownList) == 0 {
		return nil
	}
	return this.knownList[rand.Intn(len(this.knownList))]
}

func (this *KnownNodesType) GetBestNotConnectedKnownNode() *known_node.KnownNodeScored {
	this.knownNotConnectedMaxHeapMutex.RLock()
	top, _ := this.knownNotConnectedMaxHeap.GetTop()
	this.knownNotConnectedMaxHeapMutex.RUnlock()
	if top == nil {
		return nil
	}
	if top.Score == float64(known_node.KNOWN_KNODE_SCORE_MINIMUM) {
		return this.GetRandomKnownNode()
	}
	knownNode, _ := this.knownMap.Load(string(top.Key))
	return knownNode
}

func (this *KnownNodesType) IncreaseKnownNodeScore(knownNode *known_node.KnownNodeScored, delta int32, isServer bool) bool {
	update, score := knownNode.IncreaseScore(delta, isServer)
	if update {
		this.knownNotConnectedMaxHeapMutex.Lock()
		defer this.knownNotConnectedMaxHeapMutex.Unlock()
		this.knownNotConnectedMaxHeap.Update(float64(score), []byte(knownNode.URL))
	}
	return update
}

func (this *KnownNodesType) DecreaseKnownNodeScore(knownNode *known_node.KnownNodeScored, delta int32, isServer bool) (bool, bool) {
	update, removed, score := knownNode.DecreaseScore(delta, isServer)
	if removed {
		this.RemoveKnownNode(knownNode)
	}
	if update || removed {
		this.knownNotConnectedMaxHeapMutex.Lock()
		defer this.knownNotConnectedMaxHeapMutex.Unlock()
		this.knownNotConnectedMaxHeap.DeleteByKey([]byte(knownNode.URL))
		if !removed {
			this.knownNotConnectedMaxHeap.Insert(float64(score), []byte(knownNode.URL))
		}
	}
	return update, removed
}

func (this *KnownNodesType) MarkKnownNodeConnected(knownNode *known_node.KnownNodeScored) {
	this.knownNotConnectedMaxHeapMutex.Lock()
	defer this.knownNotConnectedMaxHeapMutex.Unlock()
	this.knownNotConnectedMaxHeap.DeleteByKey([]byte(knownNode.URL))
}

func (this *KnownNodesType) MarkKnownNodeDisconnected(knownNode *known_node.KnownNodeScored) {
	this.knownNotConnectedMaxHeapMutex.Lock()
	defer this.knownNotConnectedMaxHeapMutex.Unlock()
	this.knownNotConnectedMaxHeap.Update(float64(atomic.LoadInt32(&knownNode.Score)), []byte(knownNode.URL))
}

func (this *KnownNodesType) AddKnownNode(url string, isSeed bool) (*known_node.KnownNodeScored, error) {

	if url == "" {
		return nil, errors.New("url is empty")
	}

	if atomic.LoadInt32(&this.knownCount) > network_config.NETWORK_KNOWN_NODES_LIMIT {
		return nil, errors.New("Too many nodes already in the list")
	}

	if banned_nodes.BannedNodes.IsBanned(url) {
		return nil, errors.New("url is banned")
	}

	knownNode := &known_node.KnownNodeScored{
		KnownNode: known_node.KnownNode{
			URL:    url,
			IsSeed: isSeed,
		},
		Score: 0,
	}

	if _, exists := this.knownMap.LoadOrStore(url, knownNode); exists {
		return nil, errors.New("Already exists")
	}

	this.knownListMutex.Lock()
	this.knownList = append(this.knownList, knownNode)
	this.knownListMutex.Unlock()

	atomic.AddInt32(&this.knownCount, +1)

	if _, ok := connected_nodes.ConnectedNodes.AllAddresses.Load(url); !ok {
		this.knownNotConnectedMaxHeapMutex.Lock()
		this.knownNotConnectedMaxHeap.Update(0, []byte(url))
		this.knownNotConnectedMaxHeapMutex.Unlock()
	}

	return knownNode, nil
}

func (this *KnownNodesType) RemoveKnownNode(knownNode *known_node.KnownNodeScored) {

	if _, exists := this.knownMap.LoadAndDelete(knownNode.URL); exists {

		this.knownNotConnectedMaxHeapMutex.Lock()
		this.knownNotConnectedMaxHeap.DeleteByKey([]byte(knownNode.URL))
		this.knownNotConnectedMaxHeapMutex.Unlock()

		this.knownListMutex.Lock()
		defer this.knownListMutex.Unlock()
		for i, knownNode2 := range this.knownList {
			if knownNode2 == knownNode {
				this.knownList[i] = this.knownList[len(this.knownList)-1]
				this.knownList = this.knownList[:len(this.knownList)-1]
				atomic.AddInt32(&this.knownCount, -1)
				return
			}
		}
	}

}

func (this *KnownNodesType) Reset(urls []string, isSeed bool) (err error) {

	this.knownNotConnectedMaxHeapMutex.Lock()
	defer this.knownNotConnectedMaxHeapMutex.Unlock()

	this.knownListMutex.Lock()
	defer this.knownListMutex.Unlock()

	this.knownList = []*known_node.KnownNodeScored{}
	this.knownNotConnectedMaxHeap.Reset()
	atomic.StoreInt32(&this.knownCount, 0)

	changes := true
	for changes {
		changes = false
		this.knownMap.Range(func(key string, value *known_node.KnownNodeScored) bool {
			this.knownMap.Delete(key)
			changes = true
			return true
		})
	}

	for _, url := range urls {

		if atomic.LoadInt32(&this.knownCount) > network_config.NETWORK_KNOWN_NODES_LIMIT {
			return
		}

		if banned_nodes.BannedNodes.IsBanned(url) {
			continue
		}

		knownNode := &known_node.KnownNodeScored{
			KnownNode: known_node.KnownNode{
				URL:    url,
				IsSeed: isSeed,
			},
			Score: 0,
		}

		this.knownMap.LoadOrStore(url, knownNode)
		this.knownList = append(this.knownList, knownNode)
		if err = this.knownNotConnectedMaxHeap.Update(float64(knownNode.Score), []byte(url)); err != nil {
			return
		}

		atomic.AddInt32(&this.knownCount, 1)

	}
	return
}

func init() {
	KnownNodes = &KnownNodesType{
		&generics.Map[string, *known_node.KnownNodeScored]{},
		make([]*known_node.KnownNodeScored, 0),
		sync.RWMutex{},
		min_max_heap.NewMaxMemoryHeap(),
		sync.RWMutex{},
		0,
	}
}
