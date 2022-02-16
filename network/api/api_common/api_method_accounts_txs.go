package api_common

import (
	"errors"
	"net/http"
	"pandora-pay/config"
	"pandora-pay/network/api/api_common/api_types"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type APIAccountTxsRequest struct {
	api_types.APIAccountBaseRequest
	Next uint64 `json:"next,omitempty" msgpack:"next,omitempty"`
}

type APIAccountTxsReply struct {
	Count uint64   `json:"count,omitempty" msgpack:"count,omitempty"`
	Txs   [][]byte `json:"txs,omitempty" msgpack:"txs,omitempty"`
}

func (api *APICommon) GetAccountTxs(r *http.Request, args *APIAccountTxsRequest, reply *APIAccountTxsReply) (err error) {

	publicKey, err := args.GetPublicKey(true)
	if err != nil {
		return
	}

	publicKeyStr := string(publicKey)

	return store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		data := reader.Get("addrTxsCount:" + publicKeyStr)
		if data == nil {
			return nil
		}

		if reply.Count, err = strconv.ParseUint(string(data), 10, 64); err != nil {
			return
		}

		if args.Next > reply.Count {
			args.Next = reply.Count
		}

		index := uint64(0)
		if args.Next > config.API_ACCOUNT_MAX_TXS {
			index = args.Next - config.API_ACCOUNT_MAX_TXS
		}

		reply.Txs = make([][]byte, args.Next-index)
		for i := index; i < args.Next; i++ {
			hash := reader.Get("addrTx:" + publicKeyStr + ":" + strconv.FormatUint(i, 10))
			if hash == nil {
				return errors.New("Error reading address transaction")
			}
			reply.Txs[args.Next-i-1] = hash
		}

		return
	})
}
