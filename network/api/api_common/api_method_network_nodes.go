package api_common

import (
	"math/rand"
	"net/http"
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

type GetNetworkNodesListReply struct {
	Nodes []*GetNetworkNodesListNode `json:"nodes"`
}

func (api *APICommon) GetList(reply *GetNetworkNodesListReply) (err error) {

	now := time.Now()
	if now.After(api.temporaryListCreation.Load().(time.Time)) {

		api.temporaryListCreation.Store(now.Add(time.Minute * 1))

		knownList := api.knownNodes.GetList()

		count := config.NETWORK_KNOWN_NODES_LIST_RETURN
		if count > len(knownList) {
			count = len(knownList)
		}

		index := 0
		newTemporaryList := &GetNetworkNodesListReply{
			Nodes: make([]*GetNetworkNodesListNode, count),
		}

		includedMap := make(map[string]bool)

		//1st my address
		if config.NETWORK_ADDRESS_URL_STRING != "" {
			newTemporaryList.Nodes[0] = &GetNetworkNodesListNode{
				config.NETWORK_ADDRESS_URL_STRING,
				3000,
			}
			index = 1
			includedMap[config.NETWORK_ADDRESS_URL_STRING] = true
		}

		//50% top
		maxHeap := min_max_heap.NewHeapMemory(func(a, b float64) bool {
			return b < a
		})

		allKnowNodes := map[string]*known_nodes.KnownNodeScored{}

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
				newTemporaryList.Nodes[index] = &GetNetworkNodesListNode{
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

	list := api.temporaryList.Load().(*GetNetworkNodesListReply)
	*reply = *list
	return
}

func (api *APICommon) NetworkNodes(r *http.Request, args *struct{}, reply *GetNetworkNodesListReply) error {
	return api.GetList(reply)
}

func (api *APICommon) GetNetworkNodes_http(values url.Values) (interface{}, error) {
	reply := &GetNetworkNodesListReply{}
	return reply, api.NetworkNodes(nil, nil, reply)
}

func (api *APICommon) GetNetworkNodes_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &GetNetworkNodesListReply{}
	return reply, api.NetworkNodes(nil, nil, reply)
}
