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
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"strconv"
	"strings"
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

func (api *API) getBlockComplete(values *url.Values) (interface{}, error) {

	request := &api_types.APIBlockCompleteRequest{api_types.APIHeightHash{0, nil}, api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}
	if err := request.ImportFromValues(values); err != nil {
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
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlock(request)
}

func (api *API) getBlockInfo(values *url.Values) (interface{}, error) {

	request := &api_types.APIBlockInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockInfo(request)
}

func (api *API) getAssetInfo(values *url.Values) (interface{}, error) {

	request := &api_types.APIAssetInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAssetInfo(request)
}

func (api *API) getTxInfo(values *url.Values) (interface{}, error) {

	request := &api_types.APITransactionInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetTxInfo(request)
}

func (api *API) getTxPreview(values *url.Values) (interface{}, error) {

	request := &api_types.APITransactionInfoRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetTxPreview(request)
}

func (api *API) getTx(values *url.Values) (interface{}, error) {

	request := &api_types.APITransactionRequest{api_types.APIHeightHash{0, nil}, api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}

	if err := request.ImportFromValues(values); err != nil {
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

	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAccount(request)
}

func (api *API) getAccountTxs(values *url.Values) (interface{}, error) {

	request := &api_types.APIAccountTxsRequest{}

	var err error
	if values.Get("next") != "" {
		if request.Next, err = strconv.ParseUint(values.Get("next"), 10, 64); err != nil {
			return nil, err
		}
	}

	if err = request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAccountTxs(request)
}

func (api *API) getAccountMempool(values *url.Values) (interface{}, error) {

	request := &api_types.APIAccountBaseRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAccountMempool(request)
}

func (api *API) getAccountMempoolNonce(values *url.Values) (interface{}, error) {
	request := &api_types.APIAccountBaseRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAccountMempoolNonce(request)
}

func (api *API) getAsset(values *url.Values) (interface{}, error) {
	request := &api_types.APIAssetRequest{ReturnType: api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAsset(request)
}

func (api *API) getAssetFeeLiquidity(values *url.Values) (interface{}, error) {
	request := &api_types.APIAssetFeeLiquidityFeeRequest{}
	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAssetFeeLiquidity(request)
}

func (api *API) getAccountsCount(values *url.Values) (interface{}, error) {

	var assetId []byte
	var err error

	if values.Get("asset") != "" {
		if assetId, err = hex.DecodeString(values.Get("asset")); err != nil {
			return nil, err
		}
	}

	return api.apiCommon.GetAccountsCount(assetId)
}

func (api *API) getAccountsKeysByIndex(values *url.Values) (interface{}, error) {

	request := &api_types.APIAccountsKeysByIndexRequest{}

	if values.Get("encodeAddresses") == "1" {
		request.EncodeAddresses = true
	}

	if err := request.ImportFromValues(values); err != nil {
		return nil, err
	}

	return api.apiCommon.GetAccountsKeysByIndex(request)
}

func (api *API) getAccountsByKeys(values *url.Values) (interface{}, error) {

	var err error

	request := &api_types.APIAccountsByKeysRequest{ReturnType: api_types.GetReturnType(values.Get("type"), api_types.RETURN_JSON)}

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

	return api.apiCommon.GetAccountsByKeys(request)
}

func (api *API) getMempool(values *url.Values) (interface{}, error) {
	request := &api_types.APIMempoolRequest{}

	var err error
	if values.Get("chainHash") != "" {
		if request.ChainHash, err = hex.DecodeString(values.Get("chainHash")); err != nil {
			return nil, err
		}
	}
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

	return api.apiCommon.PostMempoolInsert(tx, advanced_connection_types.UUID_ALL)
}

func CreateAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain) *API {

	api := API{
		chain:     chain,
		apiStore:  apiStore,
		apiCommon: apiCommon,
	}

	api.GetMap = map[string]func(values *url.Values) (interface{}, error){
		"":                       api.getInfo,
		"chain":                  api.getBlockchain,
		"blockchain":             api.getBlockchain,
		"sync":                   api.getBlockchainSync,
		"ping":                   api.getPing,
		"block":                  api.getBlock,
		"block-hash":             api.getBlockHash,
		"block-complete":         api.getBlockComplete,
		"tx":                     api.getTx,
		"tx-hash":                api.getTxHash,
		"account":                api.getAccount,
		"accounts/count":         api.getAccountsCount,
		"accounts/keys-by-index": api.getAccountsKeysByIndex,
		"accounts/by-keys":       api.getAccountsByKeys,
		"asset":                  api.getAsset,
		"asset/fee-liquidity":    api.getAssetFeeLiquidity,
		"mem-pool":               api.getMempool,
		"mem-pool/tx-exists":     api.getMempoolExists,
		"mem-pool/new-tx":        api.postMempoolInsert,
	}

	if config.SEED_WALLET_NODES_INFO {
		api.GetMap["asset-info"] = api.getAssetInfo
		api.GetMap["block-info"] = api.getBlockInfo
		api.GetMap["tx-info"] = api.getTxInfo
		api.GetMap["tx-preview"] = api.getTxPreview
		api.GetMap["account/txs"] = api.getAccountTxs
		api.GetMap["account/mem-pool"] = api.getAccountMempool
		api.GetMap["account/mem-pool-nonce"] = api.getAccountMempoolNonce
	}

	if api.apiCommon.APICommonFaucet != nil {
		api.GetMap["faucet/info"] = api.apiCommon.APICommonFaucet.GetFaucetInfoHttp
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = api.apiCommon.APICommonFaucet.GetFaucetCoinsHttp
		}
	}

	if api.apiCommon.APIDelegatesNode != nil {
		api.GetMap["delegates/info"] = api.apiCommon.APIDelegatesNode.GetDelegatesInfoHttp
		api.GetMap["delegates/ask"] = api.apiCommon.APIDelegatesNode.GetDelegatesAskHttp
	}

	return &api
}
