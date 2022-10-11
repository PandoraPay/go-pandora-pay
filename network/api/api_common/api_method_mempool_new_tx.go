package api_common

import (
	"context"
	"net/http"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/helpers/advanced_buffers"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
)

type APIMempoolNewTxRequest struct {
	Tx helpers.Base64 `json:"tx" msgpack:"tx"`
}

type APIMempoolNewTxReply struct {
	Result bool `json:"result" msgpack:"result"`
}

func (api *APICommon) mempoolNewTx(args *APIMempoolNewTxRequest, reply *APIMempoolNewTxReply, exceptSocketUUID advanced_connection_types.UUID) (err error) {

	hash := cryptography.SHA3(args.Tx)

	//it needs to compute  tx.Bloom.HashStrx
	hashStr := string(hash)

	if api.mempool.Txs.Exists(hashStr) {
		(*reply).Result = true
		return nil
	}

	mempoolProcessedThisBlock := api.mempoolProcessedThisBlock.Load()
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.LoadOrStore(hashStr, &mempoolNewTxReply{make(chan struct{}), false, nil})

	if loaded {
		<-processedAlreadyFound.wait
		reply.Result = processedAlreadyFound.result
		return processedAlreadyFound.err
	}

	defer func() {
		if errReturned := recover(); errReturned != nil {
			err = errReturned.(error)
		}
		processedAlreadyFound.err = err
		processedAlreadyFound.result = reply.Result
		close(processedAlreadyFound.wait)
	}()

	tx := &transaction.Transaction{}
	if err = tx.Deserialize(advanced_buffers.NewBufferReader(args.Tx)); err != nil {
		return err
	}

	if err = api.txsValidator.ValidateTx(tx); err != nil {
		return
	}

	if err = api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, true, false, exceptSocketUUID, context.Background()); err != nil {
		return
	}

	(*reply).Result = true
	return
}

func (api *APICommon) MempoolNewTx(r *http.Request, args *APIMempoolNewTxRequest, reply *APIMempoolNewTxReply) error {
	return api.mempoolNewTx(args, reply, advanced_connection_types.UUID_ALL)
}
