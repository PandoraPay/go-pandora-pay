package api_websockets

import (
	"encoding/json"
	"errors"
	"pandora-pay/addresses"
	"pandora-pay/blockchain"
	"pandora-pay/blockchain/accounts/account"
	"pandora-pay/blockchain/tokens/token"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api-common"
	"pandora-pay/network/websocks/connection"
	"sync"
)

type APIWebsockets struct {
	GetMap                 map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error)
	chain                  *blockchain.Blockchain
	mempool                *mempool.Mempool
	apiCommon              *api_common.APICommon
	apiStore               *api_common.APIStore
	mempoolDownloadPending sync.Map //string
}

func (api *APIWebsockets) getHandshake(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return json.Marshal(&APIHandshake{config.NAME, config.VERSION, string(config.NETWORK_SELECTED)})
}

func (api *APIWebsockets) getBlockchain(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	data, err := api.apiCommon.GetBlockchain()
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (api *APIWebsockets) getBlockchainSync(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	data, err := api.apiCommon.GetBlockchainSync()
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (api *APIWebsockets) getInfo(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	data, err := api.apiCommon.GetInfo()
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (api *APIWebsockets) getPing(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	data, err := api.apiCommon.GetPing()
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (api *APIWebsockets) getHash(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	blockHeight := APIBlockHeight(0)
	if err := json.Unmarshal(values, &blockHeight); err != nil {
		return nil, err
	}
	out, err := api.apiCommon.GetBlockHash(blockHeight)
	if err != nil {
		return nil, err
	}
	return out.([]byte), nil
}

func (api *APIWebsockets) getBlock(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	blockHeight := APIBlockHeight(0)
	if err := json.Unmarshal(values, &blockHeight); err != nil {
		return nil, err
	}

	data, err := api.apiCommon.GetBlock(blockHeight, nil)
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (api *APIWebsockets) getBlockComplete(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	blockHeight := APIBlockHeight(0)
	if err := json.Unmarshal(values, &blockHeight); err != nil {
		return nil, err
	}

	out, err := api.apiCommon.GetBlockComplete(blockHeight, nil, 1)
	if err != nil {
		return nil, err
	}
	return out.([]byte), nil
}

func (api *APIWebsockets) getAccount(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	address, err := addresses.DecodeAddr(string(values))
	if err != nil {
		return nil, err
	}
	data, err := api.apiCommon.GetAccount(address, nil)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return data.(*account.Account).SerializeToBytes(), nil
}

func (api *APIWebsockets) getToken(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	data, err := api.apiCommon.GetToken(values)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return data.(*token.Token).SerializeToBytes(), nil
}

func (api *APIWebsockets) getMempool(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	data, err := api.apiCommon.GetMempool()
	if err != nil {
		return nil, err
	}
	return json.Marshal(data)
}

func (api *APIWebsockets) getMempoolInsert(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(values)); err != nil {
		return
	}

	var inserted bool
	if inserted, err = api.mempool.AddTxToMemPool(tx, api.chain.GetChainData().Height, true); err != nil {
		return
	}

	if inserted {
		out = []byte{1}
	} else {
		out = []byte{0}
	}

	return
}

func (api *APIWebsockets) getTx(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	data, err := api.apiCommon.GetTx(values, 1)
	if err != nil {
		return nil, err
	}

	return json.Marshal(data)
}

func (api *APIWebsockets) getMempoolExists(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	if len(values) != 32 {
		return nil, errors.New("Invalid hash")
	}
	if api.mempool.Exists(values) != nil {
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

	if api.mempool.Exists(values) != nil {
		out = []byte{1}
	} else {

		if _, loaded := api.mempoolDownloadPending.LoadOrStore(hashStr, true); loaded == true {
			out = []byte{1}
			return
		}

		result := conn.SendAwaitAnswer([]byte("tx"), values)

		if result.Out != nil && result.Err == nil {

			data := &api_common.APITransactionSerialized{}
			if err = json.Unmarshal(result.Out, data); err != nil {
				return
			}

			tx := &transaction.Transaction{}
			if err = tx.Deserialize(helpers.NewBufferReader(data.Tx)); err != nil {
				return
			}

			var output interface{}
			if output, err = api.apiCommon.PostMempoolInsert(tx); err != nil {
				return
			}
			inserted := output.(bool)

			if inserted {
				out = []byte{1}
			}

		}

		api.mempoolDownloadPending.Delete(hashStr)
	}

	return
}

func CreateWebsocketsAPI(apiStore *api_common.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain, mempool *mempool.Mempool) *APIWebsockets {

	api := APIWebsockets{
		chain:                  chain,
		apiStore:               apiStore,
		apiCommon:              apiCommon,
		mempool:                mempool,
		mempoolDownloadPending: sync.Map{},
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
		"account":            api.getAccount,
		"token":              api.getToken,
		"mem-pool":           api.getMempool,
		"mem-pool/tx-exists": api.getMempoolExists,
		"mem-pool/new-tx":    api.getMempoolInsert,
		"mem-pool/new-tx-id": api.getMempoolTxInsert,
	}

	return &api
}
