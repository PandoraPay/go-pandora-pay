package api_common

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/blockchain/data_storage/accounts/account"
	"pandora-pay/blockchain/data_storage/registrations/registration"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"strings"
)

type APIAccountsByKeysRequest struct {
	Keys           []*api_types.APIAccountBaseRequest `json:"keys,omitempty"`
	Asset          helpers.HexBytes                   `json:"asset,omitempty"`
	IncludeMempool bool                               `json:"includeMempool,omitempty"`
	ReturnType     api_types.APIReturnType            `json:"returnType,omitempty"`
}

type APIAccountsByKeys struct {
	Acc           []*account.Account           `json:"acc,omitempty"`
	AccSerialized []helpers.HexBytes           `json:"accSerialized,omitempty"`
	Reg           []*registration.Registration `json:"registration,omitempty"`
	RegSerialized []helpers.HexBytes           `json:"registrationSerialized,omitempty"`
}

func (api *APICommon) getAccountsByKeys(request *APIAccountsByKeysRequest) ([]byte, error) {

	publicKeys := make([][]byte, len(request.Keys))
	var err error

	for i, key := range request.Keys {
		if publicKeys[i], err = key.GetPublicKey(); err != nil {
			return nil, err
		}
	}

	out, err := api.ApiStore.openLoadAccountsByKeys(publicKeys, request.Asset)
	if err != nil {
		return nil, err
	}

	if request.IncludeMempool {
		balancesInit := make([][]byte, len(publicKeys))
		for i, acc := range out.Acc {
			if acc != nil {
				balancesInit[i] = helpers.SerializeToBytes(acc.Balance)
			}
		}
		if balancesInit, err = api.mempool.GetZetherBalanceMultiple(publicKeys, balancesInit, request.Asset); err != nil {
			return nil, err
		}
		for i, acc := range out.Acc {
			if balancesInit[i] != nil {
				if err = acc.Balance.Deserialize(helpers.NewBufferReader(balancesInit[i])); err != nil {
					return nil, err
				}
			}
		}
	}

	if request.ReturnType == api_types.RETURN_SERIALIZED {
		out.AccSerialized = make([]helpers.HexBytes, len(out.Acc))
		for i, acc := range out.Acc {
			if acc != nil {
				out.AccSerialized[i] = helpers.SerializeToBytes(acc)
			}
		}
		out.Acc = nil

		out.RegSerialized = make([]helpers.HexBytes, len(out.Reg))
		for i, reg := range out.Reg {
			if reg != nil {
				out.RegSerialized[i] = helpers.SerializeToBytes(reg)
			}
		}
		out.Reg = nil
	}
	return json.Marshal(out)
}

func (api *APICommon) GetAccountsByKeys_http(values *url.Values) (interface{}, error) {

	var err error

	request := &APIAccountsByKeysRequest{ReturnType: api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}

	if values.Get("publicKeys") != "" {
		v := strings.Split(values.Get("publicKeys"), ",")
		request.Keys = make([]*api_types.APIAccountBaseRequest, len(v))
		for i := range v {
			request.Keys[i] = &api_types.APIAccountBaseRequest{}
			if request.Keys[i].PublicKey, err = hex.DecodeString(v[i]); err != nil {
				return nil, err
			}
		}
	} else if values.Get("addresses") != "" {
		v := strings.Split(values.Get("addresses"), ",")
		request.Keys = make([]*api_types.APIAccountBaseRequest, len(v))
		for i := range v {
			request.Keys[i] = &api_types.APIAccountBaseRequest{Address: v[i]}
		}
	} else {
		return nil, errors.New("parameter `publicKeys` or `addresses` are missing")
	}

	if values.Get("asset") != "" {
		if request.Asset, err = hex.DecodeString(values.Get("asset")); err != nil {
			return nil, err
		}
	}
	request.IncludeMempool = values.Get("includeMempool") == "1"

	return api.getAccountsByKeys(request)
}

func (api *APICommon) GetAccountsByKeys_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &APIAccountsByKeysRequest{nil, nil, false, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.getAccountsByKeys(request)
}
