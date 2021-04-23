package api_websockets

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain"
	block_complete "pandora-pay/blockchain/block-complete"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers"
	"pandora-pay/mempool"
	"pandora-pay/network/api/api-common"
	api_store "pandora-pay/network/api/api-store"
	"pandora-pay/network/websocks/connection"
	"sync"
)

type APIWebsockets struct {
	GetMap                 map[string]func(conn *connection.AdvancedConnection, values []byte) ([]byte, error)
	chain                  *blockchain.Blockchain
	mempool                *mempool.Mempool
	apiCommon              *api_common.APICommon
	apiStore               *api_store.APIStore
	mempoolDownloadPending sync.Map //string
}

func (api *APIWebsockets) ValidateHandshake(handshake *APIHandshake) error {
	handshake2 := *handshake
	if handshake2[2] != string(config.NETWORK_SELECTED) {
		return errors.New("Network is different")
	}
	return nil
}

func (api *APIWebsockets) getHandshake(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	handshake := APIHandshake{}
	if err := json.Unmarshal(values, &handshake); err != nil {
		return nil, err
	}
	if err := api.ValidateHandshake(&handshake); err != nil {
		return nil, err
	}
	return json.Marshal(&APIHandshake{config.NAME, config.VERSION, string(config.NETWORK_SELECTED)})
}

func (api *APIWebsockets) getBlockchain(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	data, err := api.apiCommon.GetBlockchain()
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
	var blk *api_store.BlockWithTxs
	var err error

	if err := json.Unmarshal(values, &blockHeight); err != nil {
		return nil, err
	}
	if blk, err = api.apiStore.LoadBlockWithTXsFromHeight(blockHeight); err != nil {
		return nil, err
	}
	return json.Marshal(blk)
}

func (api *APIWebsockets) getBlockComplete(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {

	blockHeight := APIBlockHeight(0)
	var blkComplete *block_complete.BlockComplete
	var err error

	if err = json.Unmarshal(values, &blockHeight); err != nil {
		return nil, err
	}
	if blkComplete, err = api.apiStore.LoadBlockCompleteFromHeight(blockHeight); err != nil {
		return nil, err
	}

	return blkComplete.Serialize(), nil
}

func (api *APIWebsockets) getMempool(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	transactions := api.mempool.GetTxsList()
	hashes := make([]helpers.HexBytes, len(transactions))
	for i, tx := range transactions {
		hashes[i] = tx.Tx.Bloom.Hash
	}
	return json.Marshal(hashes)
}

func (api *APIWebsockets) getMempoolInsert(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(values)); err != nil {
		return
	}

	if err = tx.BloomAll(); err != nil {
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

func (api *APIWebsockets) getTx(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	var data interface{}
	data, err = api.apiCommon.GetTx(values, 1)
	if err != nil {
		return
	}

	out = data.([]byte)
	return
}

func (api *APIWebsockets) getMempoolTxInsert(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

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

		if result != nil && result.Err == nil {

			tx := &transaction.Transaction{}
			if err = tx.Deserialize(helpers.NewBufferReader(result.Out)); err != nil {
				return
			}

			var data interface{}
			if data, err = api.apiCommon.PostMempoolInsert(tx); err != nil {
				return
			}

			inserted := data.(bool)

			if inserted {
				out = []byte{1}
			} else {
				out = []byte{0}
			}

		}

		api.mempoolDownloadPending.Delete(hashStr)
	}

	return
}

func CreateWebsocketsAPI(apiStore *api_store.APIStore, apiCommon *api_common.APICommon, chain *blockchain.Blockchain, mempool *mempool.Mempool) *APIWebsockets {

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
		"handshake":          api.getHandshake,
		"ping":               api.getPing,
		"block":              api.getBlock,
		"block-hash":         api.getHash,
		"block-complete":     api.getBlockComplete,
		"tx":                 api.getTx,
		"mem-pool":           api.getMempool,
		"mem-pool/new-tx":    api.getMempoolInsert,
		"mem-pool/new-tx-id": api.getMempoolTxInsert,
	}

	return &api
}
