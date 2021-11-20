package api_common

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
	"sync"
)

func (api *APICommon) MempoolNewTxId_websockets(conn *connection.AdvancedConnection, values []byte) (out []byte, err error) {

	if len(values) != 32 {
		return nil, errors.New("Invalid hash")
	}
	hashStr := string(values)

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

	var exists bool
	if exists, err = api.chain.OpenExistsTx(values); exists || err != nil {
		return
	}

	result := conn.SendJSONAwaitAnswer([]byte("tx"), &APITransactionRequest{api_types.APIHeightHash{0, values}, api_types.RETURN_SERIALIZED}, nil)
	if result.Err != nil {
		err = result.Err
		return
	}

	if result.Out == nil {
		err = errors.New("Tx was not downloaded")
		return
	}

	data := &APITransactionAnswer{}
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

	if err = api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, true, false, conn.UUID); err != nil {
		return
	}

	out = []byte{1}
	return
}
