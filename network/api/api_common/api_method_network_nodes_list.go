package api_common

import (
	"encoding/json"
	"math/rand"
	"net/url"
	"pandora-pay/network/known_nodes"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store/min_max_heap"
	"time"
)

func (api *APICommon) getList() ([]*known_nodes.KnownNode, error) {

	deadline := time.Now().Add(time.Minute * 5)
	if api.temporaryListCreation.Load().(time.Time).Before(deadline) {
		api.temporaryListCreation.Store(deadline)

		knownList := api.knownNodes.GetList()

		count := 100
		if count > len(knownList) {
			count = len(knownList)
		}

		index := 0
		newTemporaryList := make([]*known_nodes.KnownNode, count)

		//1st my address

		//50% top
		maxHeap := min_max_heap.NewHeapMemory(func(a, b float64) bool {
			return b < a
		})

		allKnowNodes := map[string]*known_nodes.KnownNode{}

		for _, knownNode := range knownList {
			if err := maxHeap.Insert(float64(knownNode.Score), []byte(knownNode.UrlStr)); err != nil {
				return nil, err
			}
			allKnowNodes[knownNode.UrlStr] = &knownNode.KnownNode
		}

		includedMap := make(map[string]bool)
		for index < count/2 {
			element, err := maxHeap.GetTop()
			if err != nil {
				return nil, err
			}
			if element != nil {
				newTemporaryList[index] = allKnowNodes[string(element.Key)]
				includedMap[string(element.Key)] = true
				index += 1
			}
		}

		//50% random
		for index < count {
			for {
				element := knownList[rand.Intn(len(knownList))]
				if includedMap[element.UrlStr] {
					element = nil
				} else {
					includedMap[element.UrlStr] = true
					newTemporaryList[index] = &element.KnownNode
					index += 1
					break
				}
			}
		}

		api.temporaryList.Store(newTemporaryList)
	}

	return api.temporaryList.Load().([]*known_nodes.KnownNode), nil
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
