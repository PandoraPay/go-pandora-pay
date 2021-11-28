package api_common

import (
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"pandora-pay/cryptography"
	"pandora-pay/network/websocks/connection"
)

func (api *APICommon) MempoolExists(r *http.Request, args *[]byte, reply *[]byte) error {
	if len(*args) != cryptography.HashSize {
		return errors.New("TxId must be 32 byte")
	}
	if api.mempool.Txs.Get(string(*args)) != nil {
		*reply = []byte{1}
	} else {
		*reply = []byte{0}
	}
	return nil
}

func (api *APICommon) GetMempoolExists_http(values url.Values) (interface{}, error) {
	hash, err := hex.DecodeString(values.Get("hash"))
	if err != nil {
		return nil, err
	}
	out := []byte{}
	return out, api.MempoolExists(nil, &hash, &out)
}

func (api *APICommon) GetMempoolExists_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {
	out := []byte{}
	return out, api.MempoolExists(nil, &values, &out)
}
