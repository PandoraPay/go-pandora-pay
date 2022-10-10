package api_common

import (
	"bytes"
	"context"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers/advanced_buffers"
	"pandora-pay/network/websocks/connection"
)

func (api *APICommon) mempoolNewTxIdProcess(conn *connection.AdvancedConnection, hash []byte, reply *APIMempoolNewTxReply) (err error) {

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

	result, err := connection.SendJSONAwaitAnswer[APITxRawReply](conn, []byte("tx-raw"), &APITxRawRequest{0, hash}, nil, 0)
	if err != nil {
		closeConnection = true
		return
	}

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(advanced_buffers.NewBufferReader(result.Tx)); err != nil {
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

func (api *APICommon) MempoolNewTxId(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &APIMempoolNewTxReply{}
	return reply, api.mempoolNewTxIdProcess(conn, values, reply)
}
