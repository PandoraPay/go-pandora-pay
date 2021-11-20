package api_common

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/addresses"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
	"strconv"
	"strings"
)

type APIAccountsKeysByIndexRequest struct {
	Indexes         []uint64         `json:"indexes"`
	Asset           helpers.HexBytes `json:"asset"`
	EncodeAddresses bool             `json:"encodeAddresses"`
}

type APIAccountsKeysByIndexAnswer struct {
	PublicKeys []helpers.HexBytes `json:"publicKeys,omitempty"`
	Addresses  []string           `json:"addresses,omitempty"`
}

func (self *APIAccountsKeysByIndexRequest) ImportFromValues(values *url.Values) (err error) {

	if values.Get("indexes") != "" {
		v := strings.Split(values.Get("indexes"), ",")
		self.Indexes = make([]uint64, len(v))
		for i := 0; i < len(v); i++ {
			if self.Indexes[i], err = strconv.ParseUint(v[i], 10, 64); err != nil {
				return err
			}
		}
	} else {
		return errors.New("parameter `indexes` is missing")
	}

	if values.Get("asset") != "" {
		if self.Asset, err = hex.DecodeString(values.Get("asset")); err != nil {
			return err
		}
	}

	return
}

func (api *APICommon) getAccountsKeysByIndex(request *APIAccountsKeysByIndexRequest) ([]byte, error) {
	out, err := api.ApiStore.openLoadAccountsKeysByIndex(request.Indexes, request.Asset)
	if err != nil {
		return nil, err
	}

	answer := &APIAccountsKeysByIndexAnswer{}
	if !request.EncodeAddresses {
		answer.PublicKeys = out
	} else {
		answer.Addresses = make([]string, len(out))
		for i, publicKey := range out {
			addr, err := addresses.CreateAddr(publicKey, nil, 0, nil)
			if err != nil {
				return nil, err
			}
			answer.Addresses[i] = addr.EncodeAddr()
		}
		answer.PublicKeys = nil
	}
	return json.Marshal(answer)
}

func (api *APICommon) GetAccountsKeysByIndex_http(values *url.Values) (interface{}, error) {

	request := &APIAccountsKeysByIndexRequest{}

	if values.Get("encodeAddresses") == "1" {
		request.EncodeAddresses = true
	}

	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.getAccountsKeysByIndex(request)
}

func (api *APICommon) GetAccountsKeysByIndex_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &APIAccountsKeysByIndexRequest{nil, nil, false}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.getAccountsKeysByIndex(request)
}
