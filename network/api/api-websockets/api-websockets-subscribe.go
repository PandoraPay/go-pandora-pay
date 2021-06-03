package api_websockets

import (
	"encoding/json"
	api_common "pandora-pay/network/api/api-common"
	"pandora-pay/network/websocks/connection"
)

func (api *APIWebsockets) subscribeAccount(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	request := &api_common.APIAccountRequest{"", nil, api_common.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	publicKeyHash, err := request.GetPublicKeyHash()
	if err != nil {
		return
	}

	acc, err := api.apiStore.LoadAccountFromPublicKeyHash(publicKeyHash)
	if err != nil {
		return
	}

	if request.ReturnType == api_common.RETURN_SERIALIZED {
		out = acc.SerializeToBytes()
	} else {
		if out, err = json.Marshal(acc); err != nil {
			return
		}
	}

	conn.Subscriptions.AddSubscription([]byte("account"), publicKeyHash, request.ReturnType)

	return
}
