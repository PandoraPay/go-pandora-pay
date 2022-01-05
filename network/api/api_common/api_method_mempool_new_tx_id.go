package api_common

import (
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
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.Load(hashStr)
	if loaded {
		*reply = *processedAlreadyFound
		return nil
	}

	answer, loaded := api.MempoolDownloadPending.LoadOrStore(hashStr, &mempoolNewTxAnswer{make(chan struct{}), nil})

	if loaded {
		<-answer.wait
		*reply = *answer.reply
		return nil
	}

	defer func() {
		mempoolProcessedThisBlock.Store(hashStr, reply)
		answer.reply = reply
		close(answer.wait)
	}()

	if api.mempool.Txs.Exists(hashStr) {
		(*reply).Result = true
		return nil
	}

	if exists, err := api.chain.OpenExistsTx(hash); exists || err != nil {
		(*reply).Result = true
		return nil
	}

	result := conn.SendJSONAwaitAnswer([]byte("tx"), &APITransactionRequest{0, hash, api_types.RETURN_SERIALIZED}, nil)
	if result.Err != nil {
		(*reply).Error = result.Err
		return nil
	}

	if result.Out == nil {
		(*reply).Error = errors.New("Tx was not downloaded")
		return nil
	}

	data := &APITransactionAnswer{}
	if err := json.Unmarshal(result.Out, data); err != nil {
		(*reply).Error = err
		return nil
	}

	tx := &transaction.Transaction{}
	if err := tx.Deserialize(helpers.NewBufferReader(data.TxSerialized)); err != nil {
		(*reply).Error = err
		return nil
	}
	if err := tx.BloomAll(); err != nil {
		(*reply).Error = err
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
