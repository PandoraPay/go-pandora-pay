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
	"pandora-pay/network/api/api-common"
	"pandora-pay/network/api/api-common/api_types"
	"pandora-pay/network/websocks/connection"
	"sync"
)

type APIWebsockets struct {
	GetMap                                   map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error)
	chain                                    *blockchain.Blockchain
	mempool                                  *mempool.Mempool
	apiCommon                                *api_common.APICommon
	apiStore                                 *api_common.APIStore
	AccountsChangesSubscriptionNotifications *multicast.MulticastChannel //*api_common.APISubscriptionNotification
	mempoolDownloadPending                   *sync.Map                   //string
}

func (api *APIWebsockets) getHandshake(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return json.Marshal(&connection.ConnectionHandshake{config.NAME, config.VERSION, config.NETWORK_SELECTED, config.CONSENSUS})
}

func (api *APIWebsockets) getBlockchain(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.apiCommon.GetBlockchain()
}

func (api *APIWebsockets) getBlockchainSync(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.apiCommon.GetBlockchainSync()
}

func (api *APIWebsockets) getInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.apiCommon.GetInfo()
}

func (api *APIWebsockets) getPing(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.apiCommon.GetPing()
}

func (api *APIWebsockets) getHash(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := APIBlockHeight(0)
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetBlockHash(uint64(request))
}

func (api *APIWebsockets) getBlock(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APIBlockRequest{0, nil}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlock(request)
}

func (api *APIWebsockets) getBlockInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APIBlockRequest{0, nil}
	if err := json.Unmarshal(values, request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockInfo(request)
}

func (api *APIWebsockets) getBlockComplete(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	request := &api_types.APIBlockCompleteRequest{0, nil, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetBlockComplete(request)
}

func (api *APIWebsockets) getAccount(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APIAccountRequest{"", nil, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetAccount(request)
}

func (api *APIWebsockets) getTokenInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APITokenInfoRequest{nil}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}
	return api.apiCommon.GetTokenInfo(request)
}

func (api *APIWebsockets) getToken(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	request := &api_types.APITokenRequest{nil, api_types.RETURN_SERIALIZED}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return api.apiCommon.GetToken(request)
}

func (api *APIWebsockets) getMempool(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.apiCommon.GetMempool()
}

func (api *APIWebsockets) getMempoolInsert(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(values)); err != nil {
		return
	}

	inserted, err := api.mempool.AddTxToMemPool(tx, api.chain.GetChainData().Height, true)
	if err != nil {
		return
	}

	if inserted {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
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
	if api.mempool.Txs.Exists(values) != nil {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

func (api *APIWebsockets) getMempoolTxInsert(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	out = []byte{0}

	if len(values) != 32 {
		return nil, errors.New("Invalid hash")
	}
	hashStr := string(values)

	if api.mempool.Txs.Exists(values) != nil {
		out = []byte{1}
	} else {

		if _, loaded := api.mempoolDownloadPending.LoadOrStore(hashStr, true); loaded == true {
			out = []byte{1}
			return
		}

		result := conn.SendJSONAwaitAnswer([]byte("tx"), &api_types.APIBlockCompleteRequest{0, values, api_types.RETURN_SERIALIZED})

		if result.Out != nil && result.Err == nil {

			data := &api_types.APITransactionSerialized{}
			if err = json.Unmarshal(result.Out, data); err != nil {
				return
			}

			tx := &transaction.Transaction{}
			if err = tx.Deserialize(helpers.NewBufferReader(data.Tx)); err != nil {
				return
			}

			if out, err = api.apiCommon.PostMempoolInsert(tx); err != nil {
				return
			}
		}

		api.mempoolDownloadPending.Delete(hashStr)
	}

	return
}

func CreateWebsocketsAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain, mempool *mempool.Mempool) *APIWebsockets {

	api := &APIWebsockets{
		chain:                                    chain,
		apiStore:                                 apiStore,
		apiCommon:                                apiCommon,
		mempool:                                  mempool,
		AccountsChangesSubscriptionNotifications: multicast.NewMulticastChannel(),
		mempoolDownloadPending:                   &sync.Map{},
	}

	api.GetMap = map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error){
		"":                   api.getInfo,
		"chain":              api.getBlockchain,
		"sync":               api.getBlockchainSync,
		"handshake":          api.getHandshake,
		"ping":               api.getPing,
		"block":              api.getBlock,
		"block-hash":         api.getHash,
		"block-complete":     api.getBlockComplete,
		"tx":                 api.getTx,
		"tx-hash":            api.getTxHash,
		"account":            api.getAccount,
		"token":              api.getToken,
		"mem-pool":           api.getMempool,
		"mem-pool/tx-exists": api.getMempoolExists,
		"mem-pool/new-tx":    api.getMempoolInsert,
		"mem-pool/new-tx-id": api.getMempoolTxInsert,
	}

	if config.SEED_WALLET_NODES_INFO {
		api.GetMap["sub"] = api.subscribe
		api.GetMap["unsub"] = api.unsubscribe
		api.GetMap["sub/notify"] = api.subscribedNotificationReceived
		api.GetMap["token-info"] = api.getTokenInfo
		api.GetMap["block-info"] = api.getBlockInfo

	}

	return api
}
