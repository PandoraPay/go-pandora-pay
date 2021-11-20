package api_common

import (
	"encoding/hex"
	"net/url"
	"pandora-pay/network/websocks/connection"
	"strconv"
)

func (api *APICommon) getAccountsCount(hash []byte) (uint64, error) {
	return api.ApiStore.openLoadAccountsCountFromAssetId(hash)
}

func (api *APICommon) GetAccountsCount_http(values *url.Values) (interface{}, error) {

	var assetId []byte
	var err error

	if values.Get("asset") != "" {
		if assetId, err = hex.DecodeString(values.Get("asset")); err != nil {
			return nil, err
		}
	}

	return api.getAccountsCount(assetId)
}

func (api *APICommon) GetAccountsCount_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	count, err := api.getAccountsCount(values)
	if err != nil {
		return nil, err
	}
	return []byte(strconv.FormatUint(count, 10)), nil
}
