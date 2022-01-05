package api_common

import (
	"encoding/json"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
)

type APIMempoolNewTxRequest struct {
	Type byte             `json:"type,omitempty"`
	Tx   helpers.HexBytes `json:"tx,omitempty"`
}

type APIMempoolNewTxReply struct {
	Result bool  `json:"result"`
	Error  error `json:"error"`
}

type mempoolNewTxReply struct {
	wait  chan struct{}
	reply *APIMempoolNewTxReply
}

func (api *APICommon) mempoolNewTx(args *APIMempoolNewTxRequest, reply *APIMempoolNewTxReply, exceptSocketUUID advanced_connection_types.UUID) error {

	tx := &transaction.Transaction{}
	if args.Type == 0 {
		if err := tx.Deserialize(helpers.NewBufferReader(args.Tx)); err != nil {
			return err
		}
	} else if args.Type == 1 { //json
		if err := json.Unmarshal(args.Tx, tx); err != nil {
			return err
		}
	}

	//it needs to compute  tx.Bloom.HashStr
	hash := tx.HashManual()
	hashStr := string(hash)

	mempoolProcessedThisBlock := api.MempoolProcessedThisBlock.Load()
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.Load(hashStr)

	if loaded {
		*reply = *processedAlreadyFound
		return nil
	}

	answer, loaded := api.MempoolDownloadPending.LoadOrStore(hashStr, &mempoolNewTxReply{make(chan struct{}), nil})

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

	if err := tx.BloomAll(); err != nil {
		(*reply).Error = err
		return nil
	}
	if err := api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, true, false, exceptSocketUUID); err != nil {
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
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := &APIMempoolNewTxReply{}
	return reply, api.MempoolNewTx(nil, args, reply)
}

func (api *APICommon) MempoolNewTx_websockets(conn *connection.AdvancedConnection, values []byte) (out interface{}, err error) {
	args := &APIMempoolNewTxRequest{}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := &APIMempoolNewTxReply{}
	return reply, api.mempoolNewTx(args, reply, conn.UUID)
}
