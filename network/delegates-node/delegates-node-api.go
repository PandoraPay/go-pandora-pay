package delegates_node

import (
	"encoding/json"
	"net/url"
	"pandora-pay/network/websocks/connection"
)

func (api *DelegatesNode) getDelegatesInfoHttp(values *url.Values) (interface{}, error) {

	request := &DelegatesNodeInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getDelegatesInfo(request)
}

func (api *DelegatesNode) getDelegatesInfoWebsocket(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &DelegatesNodeInfoRequest{}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.getDelegatesInfo(request)
}
