package api_common

import (
	"bytes"
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/network/websocks/connection"
)

func (api *APICommon) mempoolNewTxId(conn *connection.AdvancedConnection, hash []byte, reply *APIMempoolNewTxReply) error {

	if len(hash) != 32 {
		return errors.New("Invalid hash")
	}
	hashStr := string(hash)

	mempoolProcessedThisBlock := api.MempoolProcessedThisBlock.Load()
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.LoadOrStore(hashStr, &mempoolNewTxReply{make(chan struct{}), nil})

	if loaded {
		<-processedAlreadyFound.wait
		*reply = *processedAlreadyFound.reply
		return nil
	}

	defer func() {
		processedAlreadyFound.reply = reply
		close(processedAlreadyFound.wait)
	}()

	if api.mempool.Txs.Exists(hashStr) {
		(*reply).Result = true
		return nil
	}

	closeConnection := func(reason error, close bool) {
		if close {
			conn.Close(reason.Error())
		}
		(*reply).Error = reason
	}

	result := conn.SendJSONAwaitAnswer([]byte("tx"), &APITransactionRequest{0, hash, api_types.RETURN_SERIALIZED}, nil)
	if result.Err != nil {
		closeConnection(result.Err, false)
		return nil
	}

	if result.Out == nil {
		closeConnection(result.Err, false)
		return nil
	}

	data := &APITransactionReply{}
	if err := json.Unmarshal(result.Out, data); err != nil {
		closeConnection(err, true)
		return nil
	}

	tx := &transaction.Transaction{}
	if err := tx.Deserialize(helpers.NewBufferReader(data.TxSerialized)); err != nil {
		closeConnection(err, true)
		return nil
	}
	if err := tx.BloomAll(); err != nil {
		closeConnection(err, true)
		return nil
	}

	if !bytes.Equal(tx.Bloom.Hash, hash) {
		closeConnection(errors.New("Wrong transaction"), true)
		return nil
	}

	if err := api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, true, false, conn.UUID); err != nil {
		(*reply).Error = err
		return nil
	}

	(*reply).Result = true
	return nil
}

func (api *APICommon) MempoolNewTxId_websockets(conn *connection.AdvancedConnection, values []byte) (out interface{}, err error) {
	reply := &APIMempoolNewTxReply{}
	return reply, api.mempoolNewTxId(conn, values, reply)
}
