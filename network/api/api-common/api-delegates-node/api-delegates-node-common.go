package api_delegates_node

import (
	"encoding/json"
	"net/url"
	"pandora-pay/network/websocks/connection"
)

func (api *APIDelegatesNode) GetDelegatesInfoHttp(values *url.Values) (interface{}, error) {

	request := &ApiDelegatesNodeInfoRequest{}
	return api.getDelegatesInfo(request)

}

func (api *APIDelegatesNode) GetDelegatesInfoWebsocket(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &ApiDelegatesNodeInfoRequest{}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.getDelegatesInfo(request)
}
