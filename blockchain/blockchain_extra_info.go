package blockchain

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/info"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

func removeBlockCompleteInfo(writer store_db_interface.StoreDBTransactionInterface, hash []byte, txHashes [][]byte, localTransactionChanges []*blockchain_types.BlockchainTransactionUpdate) (err error) {

	if err = writer.Delete("blockInfo_ByHash" + string(hash)); err != nil {
		return
	}

	for i, txHash := range txHashes {

		data := writer.Get("txKeys:" + string(txHash))
		if data == nil {
			return errors.New("TxKeys is missing")
		}
		keys := make([][]byte, 0)
		if err = json.Unmarshal(data, &keys); err != nil {
			return
		}

		localTransactionChanges[i].Keys = make([]*blockchain_types.BlockchainTransactionKeyUpdate, len(keys))
		for j, key := range keys {

			data = writer.Get("addrTxsCount:" + string(key))
			if data == nil {
				return errors.New("addrTxsCount: was empty")
			}

			var count uint64
			if count, err = strconv.ParseUint(string(data), 10, 64); err != nil {
				return
			}

			localTransactionChanges[i].Keys[j] = &blockchain_types.BlockchainTransactionKeyUpdate{
				key, count - 1,
			}

			count -= 1
			if err = writer.Delete("addrTx:" + string(key) + ":" + strconv.FormatUint(count, 10)); err != nil {
				return
			}
			if count == 0 {
				if err = writer.Delete("addrTxsCount:" + string(key)); err != nil {
					return
				}
			} else {
				if err = writer.Put("addrTxsCount:"+string(key), []byte(strconv.FormatUint(count, 10))); err != nil {
					return
				}
			}
		}

		if err = writer.Delete("txKeys:" + string(txHash)); err != nil {
			return
		}
	}

	return
}

func removeUnusedTransactions(writer store_db_interface.StoreDBTransactionInterface, starting, count uint64) (err error) {

	for i := starting; i < count; i++ {
		if err = writer.Delete("txHash_ByHeight" + strconv.FormatUint(i, 10)); err != nil {
			return errors.New("Error deleting unused transaction: " + err.Error())
		}
	}

	return
}

func removeTxsInfo(writer store_db_interface.StoreDBTransactionInterface, removedTxHashes map[string][]byte) (err error) {

	for txHash := range removedTxHashes {
		if err = writer.Delete("txInfo_ByHash" + txHash); err != nil {
			panic("Error deleting transaction info " + err.Error())
		}
		if err = writer.Delete("txPreview_ByHash" + txHash); err != nil {
			panic("Error deleting transaction info " + err.Error())
		}
	}

	return
}

func saveAssetsInfo(asts *assets.Assets) (err error) {

	for k, v := range asts.Committed {

		if v.Stored == "del" {
			err = asts.Tx.DeleteForcefully("assetInfo_ByHash:" + k)
		} else if v.Stored == "update" {

			ast := v.Element.(*asset.Asset)
			astInfo := &info.AssetInfo{
				Name:             ast.Name,
				Ticker:           ast.Ticker,
				DecimalSeparator: ast.DecimalSeparator,
				Description:      ast.Description,
			}
			var data []byte
			if data, err = json.Marshal(astInfo); err != nil {
				return
			}

			err = asts.Tx.Put("assetInfo_ByHash:"+k, data)
		}

		if err != nil {
			return
		}
	}

	return
}

func saveBlockCompleteInfo(writer store_db_interface.StoreDBTransactionInterface, blkComplete *block_complete.BlockComplete, transactionsCount uint64, localTransactionChanges []*blockchain_types.BlockchainTransactionUpdate) (err error) {

	var blockInfoMarshal []byte
	if blockInfoMarshal, err = json.Marshal(&info.BlockInfo{
		Hash:       blkComplete.Block.Bloom.Hash,
		KernelHash: blkComplete.Block.Bloom.KernelHash,
		Timestamp:  blkComplete.Block.Timestamp,
		Size:       blkComplete.BloomBlkComplete.Size,
		TXs:        uint64(len(blkComplete.Txs)),
		Forger:     blkComplete.Block.Forger,
	}); err != nil {
		return
	}

	if err = writer.Put("blockInfo_ByHash"+string(blkComplete.Block.Bloom.Hash), blockInfoMarshal); err != nil {
		return
	}

	for i, tx := range blkComplete.Txs {

		height := transactionsCount + uint64(i)
		indexStr := strconv.FormatUint(height, 10)
		if err = writer.Put("txHash_ByHeight"+indexStr, tx.Bloom.Hash); err != nil {
			return
		}

		var buffer []byte
		if buffer, err = json.Marshal(&info.TxInfo{
			height,
			blkComplete.Height,
			blkComplete.Timestamp,
		}); err != nil {
			return
		}
		if err = writer.Put("txInfo_ByHash"+tx.Bloom.HashStr, buffer); err != nil {
			return
		}

		var txPreview *info.TxPreview
		if txPreview, err = info.CreateTxPreviewFromTx(tx); err != nil {
			return
		}
		if buffer, err = json.Marshal(txPreview); err != nil {
			return
		}
		if err = writer.Put("txPreview_ByHash"+tx.Bloom.HashStr, buffer); err != nil {
			return
		}

		keys := tx.GetAllKeys()

		keysArray := make([][]byte, len(keys))
		c := 0
		for key := range keys {
			keysArray[c] = []byte(key)
			c += 1
		}

		var keysArrayMarshal []byte
		if keysArrayMarshal, err = json.Marshal(keysArray); err != nil {
			return
		}

		if err = writer.Put("txKeys:"+tx.Bloom.HashStr, keysArrayMarshal); err != nil {
			return
		}

		localTransactionChanges[i].Keys = make([]*blockchain_types.BlockchainTransactionKeyUpdate, len(keysArray))
		for j, key := range keysArray {

			keyStr := string(key)

			count := uint64(0)
			if data := writer.Get("addrTxsCount:" + keyStr); data != nil {
				if count, err = strconv.ParseUint(string(data), 10, 64); err != nil {
					return
				}
			}

			localTransactionChanges[i].Keys[j] = &blockchain_types.BlockchainTransactionKeyUpdate{
				key, count,
			}

			if err = writer.Put("addrTx:"+keyStr+":"+strconv.FormatUint(count, 10), tx.Bloom.Hash); err != nil {
				return
			}
			if err = writer.Put("addrTxsCount:"+keyStr, []byte(strconv.FormatUint(count+1, 10))); err != nil {
				return
			}
		}

	}

	return
}
