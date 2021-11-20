package api_common

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"sync"
)

func (api *APICommon) mempoolNewTx(tx *transaction.Transaction, exceptSocketUUID advanced_connection_types.UUID) (out []byte, err error) {

	//it needs to compute  tx.Bloom.HashStr
	hash := tx.HashManual()
	hashStr := string(hash)

	mempoolProcessedThisBlock := api.MempoolProcessedThisBlock.Load().(*sync.Map)
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.Load(hashStr)
	if loaded {
		if processedAlreadyFound != nil {
			return nil, processedAlreadyFound.(error)
		}
		return []byte{1}, nil
	}

	multicastFound, loaded := api.MempoolDownloadPending.LoadOrStore(hashStr, multicast.NewMulticastChannel())
	multicast := multicastFound.(*multicast.MulticastChannel)

	if loaded {
		if errData := <-multicast.AddListener(); errData != nil {
			return nil, errData.(error)
		}
		return []byte{1}, nil
	}

	defer func() {
		mempoolProcessedThisBlock.Store(hashStr, err)
		api.MempoolDownloadPending.Delete(hashStr)
		multicast.Broadcast(err)
	}()

	if api.mempool.Txs.Exists(hashStr) {
		return []byte{1}, nil
	}

	if err = tx.BloomAll(); err != nil {
		return
	}
	if err = api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, true, false, exceptSocketUUID); err != nil {
		return
	}

	return []byte{1}, nil
}

func (api *APICommon) MempoolNewTx_http(values *url.Values) (interface{}, error) {

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

	return api.mempoolNewTx(tx, advanced_connection_types.UUID_ALL)
}

func (api *APICommon) MempoolNewTx_websockets(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {
	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(values)); err != nil {
		return
	}

	return api.mempoolNewTx(tx, conn.UUID)
}
