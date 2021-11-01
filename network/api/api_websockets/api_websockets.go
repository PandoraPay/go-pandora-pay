package api_websockets

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api_common"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/settings"
	"strconv"
	"sync"
)

type APIWebsockets struct {
	GetMap                    map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error)
	chain                     *blockchain.Blockchain
	mempool                   *mempool.Mempool
	settings                  *settings.Settings
	apiCommon                 *api_common.APICommon
	apiStore                  *api_common.APIStore
	SubscriptionNotifications *multicast.MulticastChannel //*api_common.APISubscriptionNotification
}

func (api *APIWebsockets) getHandshake(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return json.Marshal(&connection.ConnectionHandshake{config.NAME, config.VERSION, config.NETWORK_SELECTED, config.CONSENSUS, config.NETWORK_ADDRESS_URL_STRING, nil})
}

func (api *APIWebsockets) getBlockchain(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.apiCommon.GetBlockchain(api_types.APIReturnType_RETURN_JSON)
}

func (api *APIWebsockets) getBlockchainSync(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.apiCommon.GetBlockchainSync(api_types.APIReturnType_RETURN_JSON)
}

func (api *APIWebsockets) getInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.apiCommon.GetInfo(api_types.APIReturnType_RETURN_JSON)
}

func (api *APIWebsockets) getPing(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.apiCommon.GetPing(api_types.APIReturnType_RETURN_JSON)
}

func (api *APIWebsockets) getHash(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := APIBlockHeight(0)
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetBlockHash(uint64(request))
}

func (api *APIWebsockets) getBlock(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APIBlockRequest{api_types.APIHeightHash{0, nil}, api_types.APIReturnType_RETURN_SERIALIZED}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlock(request)
}

func (api *APIWebsockets) getBlockInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APIBlockInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockInfo(request)
}

func (api *APIWebsockets) getBlockCompleteMissingTxs(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APIBlockCompleteMissingTxsRequest{api_types.APIHeightHash{0, nil}, []int{}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockCompleteMissingTxs(request)
}

func (api *APIWebsockets) getBlockComplete(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APIBlockCompleteRequest{api_types.APIHeightHash{0, nil}, api_types.APIReturnType_RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockComplete(request)
}

func (api *APIWebsockets) getAccount(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAccountRequest{api_types.APIAccountBaseRequest{"", nil}, api_types.APIReturnType_RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccount(request)
}

func (api *APIWebsockets) getAccountTxs(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAccountTxsRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccountTxs(request)
}

func (api *APIWebsockets) getAccountMempool(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAccountBaseRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccountMempool(request)
}

func (api *APIWebsockets) getAssetInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAssetInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAssetInfo(request)
}

func (api *APIWebsockets) getAsset(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAssetRequest{api_types.APIHeightHash{0, nil}, api_types.APIReturnType_RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAsset(request)
}

func (api *APIWebsockets) getAccountsCount(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	count, err := api.apiCommon.GetAccountsCount(values)
	if err != nil {
		return nil, err
	}
	return []byte(strconv.FormatUint(count, 10)), nil
}

func (api *APIWebsockets) getAccountsKeysByIndex(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAccountsKeysByIndexRequest{nil, nil, false}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccountsKeysByIndex(request)
}

func (api *APIWebsockets) getAccountsByKeys(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APIAccountsByKeysRequest{nil, nil, false, api_types.APIReturnType_RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccountsByKeys(request)
}

func (api *APIWebsockets) getMempool(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIMempoolRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetMempool(request)
}

func (api *APIWebsockets) getMempoolInsert(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	tx := &transaction.Transaction{}
	if err := tx.Deserialize(helpers.NewBufferReader(values)); err != nil {
		return nil, err
	}

	return api.apiCommon.PostMempoolInsert(tx, conn.UUID)
}

func (api *APIWebsockets) getTxInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APITransactionInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetTxInfo(request)
}

func (api *APIWebsockets) getTxPreview(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APITransactionInfoRequest{api_types.APIHeightHash{0, nil}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetTxPreview(request)
}

func (api *APIWebsockets) getTx(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APITransactionRequest{}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetTx(request)
}

func (api *APIWebsockets) getTxHash(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := APIBlockHeight(0)
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetTxHash(uint64(request))
}

func (api *APIWebsockets) getMempoolExists(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	if len(values) != 32 {
		return nil, errors.New("Invalid hash")
	}
	if api.mempool.Txs.Exists(string(values)) {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

func (api *APIWebsockets) postMempoolInsert(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {
	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(values)); err != nil {
		return
	}

	return api.apiCommon.PostMempoolInsert(tx, conn.UUID)
}

func (api *APIWebsockets) postMempoolTxIdInsert(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	if len(values) != 32 {
		return nil, errors.New("Invalid hash")
	}
	hashStr := string(values)

	mempoolProcessedThisBlock := api.apiCommon.MempoolProcessedThisBlock.Load().(*sync.Map)
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.Load(hashStr)
	if loaded {
		if processedAlreadyFound != nil {
			return nil, processedAlreadyFound.(error)
		}
		return []byte{1}, nil
	}

	multicastFound, loaded := api.apiCommon.MempoolDownloadPending.LoadOrStore(hashStr, multicast.NewMulticastChannel())
	multicast := multicastFound.(*multicast.MulticastChannel)

	if loaded {
		if errData := <-multicast.AddListener(); errData != nil {
			return nil, errData.(error)
		}
		return []byte{1}, nil
	}

	defer func() {
		mempoolProcessedThisBlock.Store(hashStr, err)
		api.apiCommon.MempoolDownloadPending.Delete(hashStr)
		multicast.Broadcast(err)
	}()

	if api.mempool.Txs.Exists(hashStr) {
		return []byte{1}, nil
	}

	var exists bool
	if exists, err = api.chain.OpenExistsTx(values); exists || err != nil {
		return
	}

	result := conn.SendJSONAwaitAnswer([]byte("tx"), &api_types.APITransactionRequest{api_types.APIHeightHash{0, values}, api_types.APIReturnType_RETURN_SERIALIZED}, nil)
	if result.Err != nil {
		err = result.Err
		return
	}

	if result.Out == nil {
		err = errors.New("Tx was not downloaded")
		return
	}

	data := &api_types.APITransaction{}
	if err = json.Unmarshal(result.Out, data); err != nil {
		return
	}

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(data.TxSerialized)); err != nil {
		return
	}
	if err = tx.BloomAll(); err != nil {
		return
	}

	if err = api.mempool.AddTxToMemPool(tx, api.chain.GetChainData().Height, false, true, false, conn.UUID); err != nil {
		return
	}

	out = []byte{1}
	return
}

func CreateWebsocketsAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain, settings *settings.Settings, mempool *mempool.Mempool) *APIWebsockets {

	api := &APIWebsockets{
		chain:                     chain,
		apiStore:                  apiStore,
		apiCommon:                 apiCommon,
		settings:                  settings,
		mempool:                   mempool,
		SubscriptionNotifications: multicast.NewMulticastChannel(),
	}

	api.GetMap = map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error){
		"":                       api.getInfo,
		"chain":                  api.getBlockchain,
		"blockchain":             api.getBlockchain,
		"sync":                   api.getBlockchainSync,
		"handshake":              api.getHandshake,
		"ping":                   api.getPing,
		"block":                  api.getBlock,
		"block-miss-txs":         api.getBlockCompleteMissingTxs,
		"block-hash":             api.getHash,
		"block-complete":         api.getBlockComplete,
		"tx":                     api.getTx,
		"tx-hash":                api.getTxHash,
		"account":                api.getAccount,
		"accounts/count":         api.getAccountsCount,
		"accounts/keys-by-index": api.getAccountsKeysByIndex,
		"accounts/by-keys":       api.getAccountsByKeys,
		"asset":                  api.getAsset,
		"mem-pool":               api.getMempool,
		"mem-pool/tx-exists":     api.getMempoolExists,
		"mem-pool/new-tx":        api.postMempoolInsert,
		"mem-pool/new-tx-id":     api.postMempoolTxIdInsert,
	}

	if config.SEED_WALLET_NODES_INFO {
		api.GetMap["sub"] = api.subscribe
		api.GetMap["unsub"] = api.unsubscribe
		api.GetMap["asset-info"] = api.getAssetInfo
		api.GetMap["block-info"] = api.getBlockInfo
		api.GetMap["tx-info"] = api.getTxInfo
		api.GetMap["tx-preview"] = api.getTxPreview
		api.GetMap["account/txs"] = api.getAccountTxs
		api.GetMap["account/mem-pool"] = api.getAccountMempool
	}

	if config.SEED_WALLET_NODES_INFO || config.CONSENSUS == config.CONSENSUS_TYPE_WALLET {
		api.GetMap["sub/notify"] = api.subscribedNotificationReceived
	}

	if api.apiCommon.APICommonFaucet != nil {
		api.GetMap["faucet/info"] = api.apiCommon.APICommonFaucet.GetFaucetInfoWebsocket
		if config.FAUCET_TESTNET_ENABLED {
			api.GetMap["faucet/coins"] = api.apiCommon.APICommonFaucet.GetFaucetCoinsWebsocket
		}
	}

	if api.apiCommon.APIDelegatesNode != nil {
		api.GetMap["delegates/info"] = api.apiCommon.APIDelegatesNode.GetDelegatesInfoWebsocket
		api.GetMap["delegates/ask"] = api.apiCommon.APIDelegatesNode.GetDelegatesAskWebsocket
	}

	return api
}
