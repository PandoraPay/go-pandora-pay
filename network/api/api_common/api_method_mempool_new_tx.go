package api_common

import (
	"encoding/json"
	"github.com/go-pg/urlstruct"
	"net/http"
	"net/url"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"sync"
)

type APIMempoolNewTxRequest struct {
	Type byte             `json:"type,omitempty"`
	Tx   helpers.HexBytes `json:"tx,omitempty"`
}

func (api *APICommon) mempoolNewTx(args *APIMempoolNewTxRequest, reply *[]byte, exceptSocketUUID advanced_connection_types.UUID) (err error) {

	tx := &transaction.Transaction{}
	if args.Type == 0 {
		err = tx.Deserialize(helpers.NewBufferReader(args.Tx))
	} else if args.Type == 1 { //json
		err = json.Unmarshal(args.Tx, tx)
	}

	if err != nil {
		return
	}

	//it needs to compute  tx.Bloom.HashStr
	hash := tx.HashManual()
	hashStr := string(hash)

	mempoolProcessedThisBlock := api.MempoolProcessedThisBlock.Load().(*sync.Map)
	processedAlreadyFound, loaded := mempoolProcessedThisBlock.Load(hashStr)
	if loaded {
		if processedAlreadyFound != nil {
			return processedAlreadyFound.(error)
		}
		*reply = []byte{1}
		return
	}

	multicastFound, loaded := api.MempoolDownloadPending.LoadOrStore(hashStr, multicast.NewMulticastChannel(false))
	multicast := multicastFound.(*multicast.MulticastChannel)

	if loaded {
		if errData := <-multicast.AddListener(); errData != nil {
			return errData.(error)
		}
		*reply = []byte{1}
		return
	}

	defer func() {
		mempoolProcessedThisBlock.Store(hashStr, err)
		api.MempoolDownloadPending.Delete(hashStr)
		multicast.Broadcast(err)
	}()

	if api.mempool.Txs.Exists(hashStr) {
		*reply = []byte{1}
		return
	}

	if err = tx.BloomAll(); err != nil {
		return
	}
	if err = api.mempool.AddTxToMempool(tx, api.chain.GetChainData().Height, false, true, false, exceptSocketUUID); err != nil {
		return
	}

	*reply = []byte{1}
	return
}

func (api *APICommon) MempoolNewTx(r *http.Request, args *APIMempoolNewTxRequest, reply *[]byte) error {
	return api.mempoolNewTx(args, reply, advanced_connection_types.UUID_ALL)
}

func (api *APICommon) MempoolNewTx_http(values url.Values) (interface{}, error) {
	args := &APIMempoolNewTxRequest{}
	if err := urlstruct.Unmarshal(nil, values, args); err != nil {
		return nil, err
	}
	reply := []byte{}
	return reply, api.MempoolNewTx(nil, args, &reply)
}

func (api *APICommon) MempoolNewTx_websockets(conn *connection.AdvancedConnection, values []byte) (out interface{}, err error) {
	args := &APIMempoolNewTxRequest{}
	if err := json.Unmarshal(values, args); err != nil {
		return nil, err
	}
	reply := []byte{}
	return reply, api.mempoolNewTx(args, &reply, conn.UUID)
}
