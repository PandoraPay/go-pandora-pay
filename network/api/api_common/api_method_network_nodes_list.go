package api_common

import (
	"encoding/json"
	"math/rand"
	"net/url"
	"pandora-pay/config"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store/min_max_heap"
	"sync/atomic"
	"time"
)

type GetNetworkNodesListNode struct {
	URL   string `json:"url"`
	Score int    `json:"score"`
}

type GetNetworkNodesListAnswer struct {
	Nodes []*GetNetworkNodesListNode `json:"nodes"`
}

func (api *APICommon) getList() (*GetNetworkNodesListAnswer, error) {

	deadline := time.Now().Add(time.Minute * 5)
	if api.temporaryListCreation.Load().(time.Time).Before(deadline) {
		api.temporaryListCreation.Store(deadline)

		knownList := api.knownNodes.GetList()

		count := 100
		if count > len(knownList) {
			count = len(knownList)
		}

		index := 0
		newTemporaryList := &GetNetworkNodesListAnswer{
			Nodes: make([]*GetNetworkNodesListNode, count),
		}

		//1st my address
		if config.NETWORK_ADDRESS_URL_STRING != "" {
			newTemporaryList.Nodes[0] = &GetNetworkNodesListNode{
				config.NETWORK_ADDRESS_URL_STRING,
				3000,
			}
			index = 1
		}

		//50% top
		maxHeap := min_max_heap.NewHeapMemory(func(a, b float64) bool {
			return b < a
		})

		allKnowNodes := map[string]*known_nodes.KnownNodeScored{}

		for _, knownNode := range knownList {
			if err := maxHeap.Insert(float64(atomic.LoadInt32(&knownNode.Score)), []byte(knownNode.URL)); err != nil {
				return nil, err
			}
			allKnowNodes[knownNode.URL] = knownNode
		}

		includedMap := make(map[string]bool)
		for index < count/2 {
			element, err := maxHeap.GetTop()
			if err != nil {
				return nil, err
			}
			if element != nil {

				node := allKnowNodes[string(element.Key)]
				newTemporaryList.Nodes[index] = &GetNetworkNodesListNode{
					node.URL,
					int(atomic.LoadInt32(&node.Score)),
				}
				includedMap[string(element.Key)] = true
				index += 1
			}
		}

		//50% random
		for index < count {
			for {
				node := knownList[rand.Intn(len(knownList))]
				if includedMap[node.URL] {
					node = nil
				} else {
					includedMap[node.URL] = true

					newTemporaryList.Nodes[index] = &GetNetworkNodesListNode{
						node.URL,
						int(atomic.LoadInt32(&node.Score)),
					}
					index += 1
					break
				}
			}
		}

		api.temporaryList.Store(newTemporaryList)
	}

	return api.temporaryList.Load().(*GetNetworkNodesListAnswer), nil
}

func (api *APICommon) getNetworkNodesList() ([]byte, error) {
	list, err := api.getList()
	if err != nil {
		return nil, err
	}
	return json.Marshal(list)
}

func (api *APICommon) GetNetworkNodesList_http(values *url.Values) (interface{}, error) {
	return api.getNetworkNodesList()
}

func (api *APICommon) GetNetworkNodesList_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.getNetworkNodesList()
}
