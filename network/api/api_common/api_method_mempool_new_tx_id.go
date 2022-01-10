package api_common

import (
	"bytes"
	"context"
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

	if api.mempool.Txs.Exists(hashStr) {
		(*reply).Result = true
		return nil
	}

	mempoolProcessedThisBlock := api.mempoolProcessedThisBlock.Load()
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.LoadOrStore(hashStr, &mempoolNewTxReply{make(chan struct{}), nil})

	if loaded {
		reply.Result = true
		return nil
	}

	defer func() {
		processedAlreadyFound.reply = reply
		close(processedAlreadyFound.wait)
	}()

	closeConnection := func(reason error, close bool) {
		if close {
			conn.Close(reason.Error())
		}
		(*reply).Error = reason
	}

	result := conn.SendJSONAwaitAnswer([]byte("tx"), &APITransactionRequest{0, hash, api_types.RETURN_SERIALIZED}, nil)
	if result.Err != nil {
		return result.Err
	}

	if result.Out == nil {
		return errors.New("Tx was not found")
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

	if err := api.txsValidator.ValidateTx(tx); err != nil {
		closeConnection(err, true)
		return nil
	}

	if !bytes.Equal(tx.Bloom.Hash, hash) {
		closeConnection(errors.New("Wrong transaction"), true)
		return nil
	}

	if err := api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, false, false, conn.UUID, context.Background()); err != nil {
		(*reply).Error = err
		return nil
	}

	(*reply).Result = true
	return nil
}

func (api *APICommon) MempoolNewTxId_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &APIMempoolNewTxReply{}
	return reply, api.mempoolNewTxId(conn, values, reply)
}
