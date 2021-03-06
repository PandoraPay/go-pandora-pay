package api_http

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api-common"
	"pandora-pay/network/api/api-common/api_types"
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

func (api *API) getFaucetInfo(values *url.Values) (interface{}, error) {
	return api.apiCommon.ApiCommonFaucet.GetFaucetInfo()
}

func (api *API) getFaucetCoins(values *url.Values) (interface{}, error) {

	request := &api_types.APIFaucetCoinsRequest{"", ""}

	if values.Get("address") != "" {
		request.Address = values.Get("address")
	} else {
		return nil, errors.New("parameter 'address' was not specified")
	}

	if values.Get("faucetToken") != "" {
		request.FaucetToken = values.Get("faucetToken")
	}

	return api.apiCommon.ApiCommonFaucet.GetFaucetCoins(request)
}

func (api *API) getBlockComplete(values *url.Values) (interface{}, error) {

	request := &api_types.APIBlockCompleteRequest{0, nil, api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}

	err := errors.New("parameter 'hash' or 'height' are missing")
	if values.Get("height") != "" {
		request.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	}
	if err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockComplete(request)
}

func (api *API) getBlockHash(values *url.Values) (interface{}, error) {

	if values.Get("height") != "" {
		height, err := strconv.ParseUint(values.Get("height"), 10, 64)
		if err != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}
		return api.apiCommon.GetBlockHash(height)
	}

	return nil, errors.New("parameter `height` is missing")
}

func (api *API) getBlock(values *url.Values) (interface{}, error) {

	request := &api_types.APIBlockRequest{}

	err := errors.New("parameter 'hash' or 'height' are missing")
	if values.Get("height") != "" {
		request.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	}
	if err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlock(request)
}

func (api *API) getBlockInfo(values *url.Values) (interface{}, error) {

	request := &api_types.APIBlockInfoRequest{}

	err := errors.New("parameter 'hash' or 'height' are missing")
	if values.Get("height") != "" {
		request.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	}
	if err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockInfo(request)
}

func (api *API) getTokenInfo(values *url.Values) (interface{}, error) {

	request := &api_types.APITokenInfoRequest{}

	err := errors.New("parameter 'hash' is missing")
	if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	}
	if err != nil {
		return nil, err
	}

	return api.apiCommon.GetTokenInfo(request)
}

func (api *API) getTxInfo(values *url.Values) (interface{}, error) {

	request := &api_types.APITransactionInfoRequest{}

	err := errors.New("parameter 'hash' or 'height' are missing")
	if values.Get("height") != "" {
		request.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	}
	if err != nil {
		return nil, err
	}

	return api.apiCommon.GetTxInfo(request)
}

func (api *API) getTx(values *url.Values) (interface{}, error) {

	request := &api_types.APITransactionRequest{0, nil, api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}

	err := errors.New("parameter 'hash' or 'height' are missing")
	if values.Get("height") != "" {
		request.Height, err = strconv.ParseUint(values.Get("height"), 10, 64)
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	}
	if err != nil {
		return nil, err
	}

	return api.apiCommon.GetTx(request)
}

func (api *API) getTxHash(values *url.Values) (interface{}, error) {
	if values.Get("height") != "" {
		height, err := strconv.ParseUint(values.Get("height"), 10, 64)
		if err != nil {
			return nil, errors.New("parameter 'height' is not a number")
		}

		return api.apiCommon.GetTxHash(height)
	}
	return nil, errors.New("parameter `height` is missing")
}

func (api *API) getAccount(values *url.Values) (interface{}, error) {
	request := &api_types.APIAccountRequest{api_types.APIAccountBaseRequest{"", nil}, api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}

	var err error
	if values.Get("address") != "" {
		request.Address = values.Get("address")
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	} else {
		err = errors.New("parameter 'address' or 'hash' was not specified")
	}
	if err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccount(request)
}

func (api *API) getAccountTxs(values *url.Values) (interface{}, error) {

	request := &api_types.APIAccountTxsRequest{}

	var err error
	if values.Get("next") != "" {
		if request.Next, err = strconv.ParseUint(values.Get("start"), 10, 64); err != nil {
			return nil, err
		}
	}

	if values.Get("address") != "" {
		request.Address = values.Get("address")
	} else if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	} else {
		err = errors.New("parameter 'address' or 'hash' was not specified")
	}
	if err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccountTxs(request)
}

func (api *API) getToken(values *url.Values) (interface{}, error) {
	request := &api_types.APITokenRequest{}
	request.ReturnType = api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)

	err := errors.New("parameter 'hash' was not specified")
	if values.Get("hash") != "" {
		request.Hash, err = hex.DecodeString(values.Get("hash"))
	}
	if err != nil {
		return nil, err
	}
	return api.apiCommon.GetToken(request)
}

func (api *API) getMempool(values *url.Values) (interface{}, error) {
	request := &api_types.APIMempoolRequest{}

	var err error
	if values.Get("page") != "" {
		if request.Page, err = strconv.Atoi(values.Get("page")); err != nil {
			return nil, err
		}
	}
	if values.Get("count") != "" {
		if request.Count, err = strconv.Atoi(values.Get("count")); err != nil {
			return nil, err
		}
	}

	return api.apiCommon.GetMempool(request)
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

	err := errors.New("parameter 'type' was not specified or is invalid")
	if values.Get("type") == "json" {
		data := values.Get("tx")
		err = json.Unmarshal([]byte(data), tx)
	} else if values.Get("type") == "binary" {
		data, err := hex.DecodeString(values.Get("tx"))
		if err != nil {
			return nil, err
		}
		err = tx.Deserialize(helpers.NewBufferReader(data))
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
		"tx-hash":            api.getTxHash,
		"account":            api.getAccount,
		"token":              api.getToken,
		"mem-pool":           api.getMempool,
		"mem-pool/tx-exists": api.getMempoolExists,
		"mem-pool/new-tx":    api.postMempoolInsert,
	}

	if config.SEED_WALLET_NODES_INFO {
		api.GetMap["token-info"] = api.getTokenInfo
		api.GetMap["block-info"] = api.getBlockInfo
		api.GetMap["tx-info"] = api.getTxInfo
		api.GetMap["account/txs"] = api.getAccountTxs
	}

	if config.NETWORK_SELECTED == config.TEST_NET_NETWORK_BYTE || config.NETWORK_SELECTED == config.DEV_NET_NETWORK_BYTE {
		api.GetMap["faucet/info"] = api.getFaucetInfo
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = api.getFaucetCoins
		}
	}

	return &api
}
