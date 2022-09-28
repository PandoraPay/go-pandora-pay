package api_common

import (
	"math/rand"
	"net/http"
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
	"pandora-pay/network/known_nodes/known_node"
	"pandora-pay/store/min_max_heap"
	"sync/atomic"
	"time"
)

type APINetworkNode struct {
	URL   string `json:"url" msgpack:"url"`
	Score int    `json:"score" msgpack:"score"`
}

type APINetworkNodesReply struct {
	Nodes []*APINetworkNode `json:"nodes" msgpack:"nodes"`
}

func (api *APICommon) GetList(reply *APINetworkNodesReply) (err error) {

	now := time.Now()
	if now.After(api.temporaryListCreation.Load()) {

		api.temporaryListCreation.Store(now.Add(time.Minute * 1))

		knownList := api.knownNodes.GetList()

		count := generics.Min(config.NETWORK_KNOWN_NODES_LIST_RETURN, len(knownList))

		index := 0
		newTemporaryList := &APINetworkNodesReply{
			Nodes: make([]*APINetworkNode, count),
		}

		includedMap := make(map[string]bool)

		//1st my address
		if config.NETWORK_WEBSOCKET_ADDRESS_URL_STRING != "" {
			newTemporaryList.Nodes[0] = &APINetworkNode{
				config.NETWORK_WEBSOCKET_ADDRESS_URL_STRING,
				3000,
			}
			index = 1
			includedMap[config.NETWORK_WEBSOCKET_ADDRESS_URL_STRING] = true
		}

		//50% top
		maxHeap := min_max_heap.NewHeapMemory(func(a, b float64) bool {
			return b < a
		})

		allKnowNodes := map[string]*known_node.KnownNodeScored{}

		for _, knownNode := range knownList {
			if err = maxHeap.Insert(float64(atomic.LoadInt32(&knownNode.Score)), []byte(knownNode.URL)); err != nil {
				return
			}
			allKnowNodes[knownNode.URL] = knownNode
		}

		for index < count/2 {
			var element *min_max_heap.HeapElement
			if element, err = maxHeap.RemoveTop(); err != nil {
				return
			}
			if element == nil {
				break
			}
			if !includedMap[string(element.Key)] {

				node := allKnowNodes[string(element.Key)]
				newTemporaryList.Nodes[index] = &APINetworkNode{
					node.URL,
					int(atomic.LoadInt32(&node.Score)),
				}
				includedMap[node.URL] = true
				index += 1
			}
		}

		//50% random
		for index < count && len(includedMap) != len(allKnowNodes) {

			for {

				node := knownList[rand.Intn(len(knownList))]
				if includedMap[node.URL] {
					node = nil
				} else {
					includedMap[node.URL] = true

					newTemporaryList.Nodes[index] = &APINetworkNode{
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

	list := api.temporaryList.Load()
	*reply = *list
	return
}

func (api *APICommon) GetNetworkNodes(r *http.Request, args *struct{}, reply *APINetworkNodesReply) error {
	return api.GetList(reply)
}
