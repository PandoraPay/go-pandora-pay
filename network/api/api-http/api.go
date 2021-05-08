package api_http

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api-common"
	"strconv"
)

type API struct {
	GetMap    map[string]func(values *url.Values) (interface{}, error)
	chain     *blockchain.Blockchain
	apiCommon *api_common.APICommon
	apiStore  *api_common.APIStore
}

func (api *API) getBlockchain(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetBlockchain()
}

func (api *API) getBlockchainSync(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetBlockchainSync()
}

func (api *API) getInfo(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetInfo()
}

func (api *API) getPing(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetPing()
}

func (api *API) getBlockComplete(values *url.Values) (out interface{}, err error) {

	var typeValue = uint8(0)
	if values.Get("type") == "1" {
		typeValue = 1
	}

	if values.Get("height") != "" {
		height, err2 := strconv.ParseUint(values.Get("height"), 10, 64)
		if err2 != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}
		out, err = api.apiCommon.GetBlockComplete(height, nil, typeValue)
	} else if values.Get("hash") != "" {
		hash, err2 := hex.DecodeString(values.Get("hash"))
		if err2 != nil {
			return nil, errors.New("parameter 'hash' is not a hex")
		}
		out, err = api.apiCommon.GetBlockComplete(0, hash, typeValue)
	} else {
		err = errors.New("parameter 'hash' or 'height' are missing")
	}

	if err != nil {
		return
	}
	if typeValue == 1 {
		return helpers.HexBytes(out.([]byte)), nil
	}

	return
}

func (api *API) getBlockHash(values *url.Values) (interface{}, error) {
	if values.Get("height") != "" {
		height, err := strconv.ParseUint(values.Get("height"), 10, 64)
		if err != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}

		out, err := api.apiCommon.GetBlockHash(height)
		if err != nil {
			return nil, err
		}
		return helpers.HexBytes(out.([]byte)), nil
	}
	return nil, errors.New("Hash parameter is missing")
}

func (api *API) getBlock(values *url.Values) (interface{}, error) {

	if values.Get("height") != "" {
		height, err := strconv.ParseUint(values.Get("height"), 10, 64)
		if err != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}
		return api.apiCommon.GetBlock(height, nil)
	}
	if values.Get("hash") != "" {
		hash, err := hex.DecodeString(values.Get("hash"))
		if err != nil {
			return nil, errors.New("parameter 'hash' was is not a valid hex number")
		}
		return api.apiCommon.GetBlock(0, hash)
	}
	return nil, errors.New("parameter 'hash' or 'height' are missing")
}

func (api *API) getTx(values *url.Values) (interface{}, error) {

	var err error
	var typeValue = uint8(0)
	if values.Get("type") == "1" {
		typeValue = 1
	}

	if values.Get("hash") != "" {
		var hash []byte
		hash, err = hex.DecodeString(values.Get("hash"))
		if err != nil {
			return nil, errors.New("parameter 'hash' was is not a valid hex number")
		}

		return api.apiCommon.GetTx(hash, typeValue)
	}

	return nil, errors.New("parameter 'hash' was not specified ")
}

func (api *API) getAccount(values *url.Values) (interface{}, error) {
	if values.Get("address") != "" {
		address, err := addresses.DecodeAddr(values.Get("address"))
		if err != nil {
			return nil, err
		}
		return api.apiCommon.GetAccount(address, nil)
	}
	if values.Get("hash") != "" {
		hash, err := hex.DecodeString(values.Get("hash"))
		if err != nil {
			return nil, err
		}
		return api.apiCommon.GetAccount(nil, hash)
	}
	return nil, errors.New("parameter 'address' or 'hash' was not specified")
}

func (api *API) getToken(values *url.Values) (interface{}, error) {
	hash, err := hex.DecodeString(values.Get("hash"))
	if err != nil {
		return nil, err
	}
	return api.apiCommon.GetToken(hash)
}

func (api *API) getMempool(values *url.Values) (interface{}, error) {
	return api.apiCommon.GetMempool()
}

func (api *API) getMempoolExists(values *url.Values) (interface{}, error) {
	hash, err := hex.DecodeString(values.Get("hash"))
	if err != nil {
		return nil, err
	}
	return api.apiCommon.GetMempoolExists(hash)
}

func (api *API) postMempoolInsert(values *url.Values) (interface{}, error) {

	tx := &transaction.Transaction{}
	var err error

	if values.Get("type") == "json" {
		data := values.Get("tx")
		err = json.Unmarshal([]byte(data), tx)
	} else if values.Get("type") == "binary" {
		data, err := hex.DecodeString(values.Get("tx"))
		if err != nil {
			return nil, err
		}
		if err = tx.Deserialize(helpers.NewBufferReader(data)); err != nil {
			return nil, err
		}
	} else {
		err = errors.New("parameter 'type' was not specified or is invalid")
	}
	if err != nil {
		return nil, err
	}

	return api.apiCommon.PostMempoolInsert(tx)
}

func CreateAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain) *API {

	api := API{
		chain:     chain,
		apiStore:  apiStore,
		apiCommon: apiCommon,
	}

	api.GetMap = map[string]func(values *url.Values) (interface{}, error){
		"":                   api.getInfo,
		"chain":              api.getBlockchain,
		"sync":               api.getBlockchainSync,
		"ping":               api.getPing,
		"block":              api.getBlock,
		"block-hash":         api.getBlockHash,
		"block-complete":     api.getBlockComplete,
		"tx":                 api.getTx,
		"account":            api.getAccount,
		"token":              api.getToken,
		"mempool":            api.getMempool,
		"mem-pool/tx-exists": api.getMempoolExists,
		"mem-pool/new-tx":    api.postMempoolInsert,
	}

	return &api
}
