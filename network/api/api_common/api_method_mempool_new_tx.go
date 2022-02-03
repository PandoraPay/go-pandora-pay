package api_common

import (
	"context"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
)

type APIMempoolNewTxRequest struct {
	Tx helpers.HexBytes `json:"tx,omitempty" msgpack:"tx,omitempty"`
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
	if err = tx.Deserialize(helpers.NewBufferReader(args.Tx)); err != nil {
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

func (api *APICommon) MempoolNewTx_http(values url.Values) (interface{}, error) {
	args := &APIMempoolNewTxRequest{}
	if err := urldecoder.Decoder.Decode(args, values); err != nil {
		return nil, err
	}
	reply := &APIMempoolNewTxReply{}
	return reply, api.MempoolNewTx(nil, args, reply)
}

func (api *APICommon) MempoolNewTx_websockets(conn *connection.AdvancedConnection, values []byte) (out interface{}, err error) {
	args := &APIMempoolNewTxRequest{}
	if err := msgpack.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIMempoolNewTxReply{}
	return reply, api.mempoolNewTx(args, reply, conn.UUID)
}
