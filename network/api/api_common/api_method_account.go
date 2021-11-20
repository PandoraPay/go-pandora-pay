package api_common

import (
	"encoding/json"
	"net/url"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/plain_accounts/plain_account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

type APIAccountRequest struct {
	api_types.APIAccountBaseRequest
	ReturnType api_types.APIReturnType `json:"returnType,omitempty"`
}

type APIAccount struct {
	Accs               []*account.Account                                      `json:"accounts,omitempty"`
	AccsSerialized     []helpers.HexBytes                                      `json:"accountsSerialized,omitempty"`
	AccsExtra          []*api_types.APISubscriptionNotificationAccountExtra    `json:"accountsExtra,omitempty"`
	PlainAcc           *plain_account.PlainAccount                             `json:"plainAccount,omitempty"`
	PlainAccSerialized helpers.HexBytes                                        `json:"plainAccountSerialized,omitempty"`
	PlainAccExtra      *api_types.APISubscriptionNotificationPlainAccExtra     `json:"plainAccountExtra,omitempty"`
	Reg                *registration.Registration                              `json:"registration,omitempty"`
	RegSerialized      helpers.HexBytes                                        `json:"registrationSerialized,omitempty"`
	RegExtra           *api_types.APISubscriptionNotificationRegistrationExtra `json:"registrationExtra,omitempty"`
}

func (api *APICommon) getAccount(request *APIAccountRequest) ([]byte, error) {

	publicKey, err := request.GetPublicKey()
	if err != nil {
		return nil, err
	}

	outAcc, err := api.ApiStore.OpenLoadAccountFromPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	if request.ReturnType == api_types.RETURN_SERIALIZED {

		outAcc.AccsSerialized = make([]helpers.HexBytes, len(outAcc.Accs))
		for i, acc := range outAcc.Accs {
			outAcc.AccsSerialized[i] = helpers.SerializeToBytes(acc)
		}
		outAcc.Accs = nil

		if outAcc.PlainAcc != nil {
			outAcc.PlainAccSerialized = helpers.SerializeToBytes(outAcc.PlainAcc)
			outAcc.PlainAcc = nil
		}
		if outAcc.Reg != nil {
			outAcc.RegSerialized = helpers.SerializeToBytes(outAcc.Reg)
			outAcc.Reg = nil
		}

	}

	return json.Marshal(outAcc)
}

func (api *APICommon) GetAccount_http(values *url.Values) (interface{}, error) {
	request := &APIAccountRequest{api_types.APIAccountBaseRequest{"", nil}, api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}

	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getAccount(request)
}

func (api *APICommon) GetAccount_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIAccountRequest{api_types.APIAccountBaseRequest{"", nil}, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.getAccount(request)
}
