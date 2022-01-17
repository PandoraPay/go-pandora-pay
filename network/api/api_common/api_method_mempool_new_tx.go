package api_common

import (
	"context"
	"github.com/vmihailenco/msgpack/v5"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/helpers/urldecoder"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
)

type APIMempoolNewTxRequest struct {
	Type byte             `json:"type,omitempty" msgpack:"type,omitempty"`
	Tx   helpers.HexBytes `json:"tx,omitempty" msgpack:"tx,omitempty"`
}

type APIMempoolNewTxReply struct {
	Result bool  `json:"result" msgpack:"result"`
	Error  error `json:"error" msgpack:"error"`
}

func (api *APICommon) mempoolNewTx(args *APIMempoolNewTxRequest, reply *APIMempoolNewTxReply, exceptSocketUUID advanced_connection_types.UUID) error {

	var hash []byte

	tx := &transaction.Transaction{}
	if args.Type == 0 {
		if err := tx.Deserialize(helpers.NewBufferReader(args.Tx)); err != nil {
			return err
		}
		hash = tx.Bloom.Hash
	} else if args.Type == 1 { //json
		if err := msgpack.Unmarshal(args.Tx, tx); err != nil {
			return err
		}
		hash = tx.HashManual()
	}

	//it needs to compute  tx.Bloom.HashStr
	hashStr := string(hash)

	if api.mempool.Txs.Exists(hashStr) {
		(*reply).Result = true
		return nil
	}

	mempoolProcessedThisBlock := api.mempoolProcessedThisBlock.Load()
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

	if err := api.txsValidator.ValidateTx(tx); err != nil {
		(*reply).Error = err
		return nil
	}

	if err := api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, false, false, exceptSocketUUID, context.Background()); err != nil {
		(*reply).Error = err
		return nil
	}

	(*reply).Result = true
	return nil
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
