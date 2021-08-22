package api_delegates_node

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/network/websocks/connection"
)

func (api *APIDelegatesNode) GetDelegatesInfoHttp(values *url.Values) (interface{}, error) {
	request := &ApiDelegatesNodeInfoRequest{}
	return api.getDelegatesInfo(request)
}

func (api *APIDelegatesNode) GetDelegatesAskHttp(values *url.Values) (interface{}, error) {
	request := &ApiDelegatesNodeAskRequest{}
	var err error
	if challengeSignature := values.Get("challengeSignature"); challengeSignature != "" {
		request.ChallengeSignature, err = hex.DecodeString(challengeSignature)
	} else {
		err = errors.New("'challengeSignature' parameter is missing")
	}
	if err != nil {
		return nil, err
	}

	return api.getDelegatesAsk(request)
}

func (api *APIDelegatesNode) GetDelegatesInfoWebsocket(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &ApiDelegatesNodeInfoRequest{}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.getDelegatesInfo(request)
}

func (api *APIDelegatesNode) GetDelegatesAskWebsocket(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &ApiDelegatesNodeAskRequest{}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.getDelegatesAsk(request)
}
