package api_common

import (
	"encoding/json"
	"errors"
	"pandora-pay/cryptography"
	"pandora-pay/helpers"
	"pandora-pay/network/websocks/connection"
	"pandora-pay/store"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

type APIBlockCompleteMissingTxsRequest struct {
	Hash       helpers.HexBytes `json:"hash,omitempty"`
	MissingTxs []int            `json:"missingTxs,omitempty"`
}

type APIBlockCompleteMissingTxs struct {
	Txs []helpers.HexBytes `json:"txs,omitempty"`
}

func (api *APICommon) getBlockCompleteMissingTxs(args *APIBlockCompleteMissingTxsRequest) (interface{}, error) {

	blockCompleteMissingTxs := &APIBlockCompleteMissingTxs{}

	if len(args.Hash) == cryptography.HashSize {
		return nil, errors.New("Invalid Block Hash")
	}
	if err := store.StoreBlockchain.DB.View(func(reader store_db_interface.StoreDBTransactionInterface) (err error) {

		heightStr := reader.Get("blockHeight_ByHash" + string(args.Hash))
		if heightStr == nil {
			return errors.New("Block was not found by hash")
		}

		var height uint64
		if height, err = strconv.ParseUint(string(heightStr), 10, 64); err != nil {
			return
		}

		data := reader.Get("blockTxs" + strconv.FormatUint(height, 10))
		if data == nil {
			return errors.New("Block not found")
		}

		txHashes := [][]byte{}
		if err = json.Unmarshal(data, &txHashes); err != nil {
			return
		}

		blockCompleteMissingTxs.Txs = make([]helpers.HexBytes, len(args.MissingTxs))
		for i, txMissingIndex := range args.MissingTxs {
			if txMissingIndex >= 0 && txMissingIndex < len(txHashes) {
				tx := reader.Get("tx:" + string(txHashes[txMissingIndex]))
				if tx == nil {
					return errors.New("Tx was not found")
				}
				blockCompleteMissingTxs.Txs[i] = tx
			}
		}

		return
	}); err != nil {
		return nil, err
	}
	return blockCompleteMissingTxs, nil
}

func (api *APICommon) GetBlockCompleteMissingTxs_websockets(conn *connection.AdvancedConnection, values []byte) (interface{}, error) {

	request := &APIBlockCompleteMissingTxsRequest{nil, []int{}}
	if err := json.Unmarshal(values, &request); err != nil {
		return nil, err
	}

	return api.getBlockCompleteMissingTxs(request)
}
