package api_common

import (
	"bytes"
	"context"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
)

func (api *APICommon) mempoolNewTxId(conn *connection.AdvancedConnection, hash []byte, reply *APIMempoolNewTxReply) (err error) {

	if len(hash) != 32 {
		return errors.New("Invalid hash")
	}
	hashStr := string(hash)

	if api.mempool.Txs.Exists(hashStr) {
		(*reply).Result = true
		return
	}

	mempoolProcessedThisBlock := api.mempoolProcessedThisBlock.Load()
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.LoadOrStore(hashStr, &mempoolNewTxReply{make(chan struct{}), false, nil})

	if loaded {
		<-processedAlreadyFound.wait
		reply.Result = processedAlreadyFound.result
		return processedAlreadyFound.err
	}

	closeConnection := false

	defer func() {
		if errReturned := recover(); errReturned != nil {
			err = errReturned.(error)
		}
		processedAlreadyFound.err = err
		processedAlreadyFound.result = reply.Result
		if closeConnection {
			mempoolProcessedThisBlock.Delete(hashStr)
		}
		close(processedAlreadyFound.wait)
	}()

	result := conn.SendJSONAwaitAnswer([]byte("tx-raw"), &APITransactionRawRequest{0, hash}, nil, 0)
	if result.Err != nil {
		closeConnection = true
		err = result.Err
		return
	}

	if result.Out == nil {
		closeConnection = true
		err = errors.New("Tx was not found")
		return
	}

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(result.Out)); err != nil {
		closeConnection = true
		return
	}

	if err = api.txsValidator.ValidateTx(tx); err != nil {
		closeConnection = true
		return
	}

	if !bytes.Equal(tx.Bloom.Hash, hash) {
		err = errors.New("Wrong transaction")
		closeConnection = true
		return
	}

	if err = api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, true, false, conn.UUID, context.Background()); err != nil {
		return
	}

	(*reply).Result = true
	return
}

func (api *APICommon) MempoolNewTxId_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &APIMempoolNewTxReply{}
	return reply, api.mempoolNewTxId(conn, values, reply)
}
