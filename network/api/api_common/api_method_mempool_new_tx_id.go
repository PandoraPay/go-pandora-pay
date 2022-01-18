package api_common

import (
	"bytes"
	"context"
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/network/api/api_common/api_types"
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
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.LoadOrStore(hashStr, &mempoolNewTxReply{make(chan struct{}), nil, nil})

	if loaded {
		<-processedAlreadyFound.wait
		*reply = *processedAlreadyFound.reply
		return processedAlreadyFound.err
	}

	closeConnection := false

	defer func() {
		processedAlreadyFound.err = err
		processedAlreadyFound.reply = reply
		if closeConnection {
			mempoolProcessedThisBlock.Delete(hashStr)
		}
		close(processedAlreadyFound.wait)
	}()

	result := conn.SendJSONAwaitAnswer([]byte("tx"), &APITransactionRequest{0, hash, api_types.RETURN_SERIALIZED}, nil, 0)
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

	data := &APITransactionReply{}
	if err = msgpack.Unmarshal(result.Out, data); err != nil {
		closeConnection = true
		return
	}

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(helpers.NewBufferReader(data.TxSerialized)); err != nil {
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

	if err = api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, false, false, conn.UUID, context.Background()); err != nil {
		return
	}

	(*reply).Result = true
	return
}

func (api *APICommon) MempoolNewTxId_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	reply := &APIMempoolNewTxReply{}
	return reply, api.mempoolNewTxId(conn, values, reply)
}
