package api_common

import (
	"encoding/hex"
	"errors"
	"net/url"
	"pandora-pay/cryptography"
	"pandora-pay/network/websocks/connection"
)

func (api *APICommon) GetMempoolExists(txId []byte) ([]byte, error) {
	if len(txId) != cryptography.HashSize {
		return nil, errors.New("TxId must be 32 byte")
	}
	if api.mempool.Txs.Get(string(txId)) != nil {
		return []byte{1}, nil
	} else {
		return []byte{0}, nil
	}
}

func (api *APICommon) GetMempoolExists_http(values *url.Values) (interface{}, error) {
	hash, err := hex.DecodeString(values.Get("hash"))
	if err != nil {
		return nil, err
	}
	return api.GetMempoolExists(hash)
}

func (api *APICommon) GetMempoolExists_websockets(conn *connection.AdvancedConnection, values []byte) ([]byte, error) {
	return api.GetMempoolExists(values)
}
