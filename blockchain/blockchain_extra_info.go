package blockchain

import (
	"encoding/json"
	"errors"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

func saveExtra(writer store_db_interface.StoreDBTransactionInterface, dataStorage *data_storage.DataStorage) (err error) {

	for k, v := range dataStorage.Asts.Committed {

		if v.Stored == "del" {
			writer.Delete("assetInfo_ByHash:" + k)
		} else if v.Stored == "update" {

			ast := v.Element.(*asset.Asset)
			astInfo := &info.AssetInfo{
				Version:          ast.Version,
				Name:             ast.Name,
				Ticker:           ast.Ticker,
				DecimalSeparator: ast.DecimalSeparator,
				Description:      ast.Description,
				Hash:             helpers.HexBytes(k),
			}
			var data []byte
			if data, err = json.Marshal(astInfo); err != nil {
				return
			}

			writer.Put("assetInfo_ByHash:"+k, data)
		}

	}

	for k, v := range dataStorage.Txs.ListChanges {

		if v.Status == "del" {
			writer.Delete("txInfo_ByHash" + k)
			writer.Delete("txPreview_ByHash" + k)

			data := writer.Get("txKeys:" + k)
			if data == nil {
				return errors.New("TxKeys is missing")
			}
			keys := make([][]byte, 0)
			if err = json.Unmarshal(data, &keys); err != nil {
				return
			}

			for j, key := range keys {

				data = writer.Get("addrTxsCount:" + string(key))
				if data == nil {
					return errors.New("addrTxsCount: was empty")
				}

				var count uint64
				if count, err = strconv.ParseUint(string(data), 10, 64); err != nil {
					return
				}

				count -= 1
				writer.Delete("addrTx:" + string(key) + ":" + strconv.FormatUint(count, 10))
				if count == 0 {
					writer.Delete("addrTxsCount:" + string(key))
				} else {
					writer.Put("addrTxsCount:"+string(key), []byte(strconv.FormatUint(count, 10)))
				}
			}

			writer.Delete("txKeys:" + k)

		} else if v.Stored == "update" {

		}

	}

	for k, v := range dataStorage.Blocks.Committed {
		if v.Stored == "del" {
			writer.Delete("blockInfo_ByHash" + k)
		} else if v.Stored == "update" {
			var buffer []byte
			if buffer, err = json.Marshal(&info.TxInfo{
				transactionsCount + uint64(i),
				blkComplete.Height,
				blkComplete.Timestamp,
			}); err != nil {
				return
			}
			writer.Put("txInfo_ByHash"+tx.Bloom.HashStr, buffer)

			var txPreview *info.TxPreview
			if txPreview, err = info.CreateTxPreviewFromTx(tx); err != nil {
				return
			}
			if buffer, err = json.Marshal(txPreview); err != nil {
				return
			}
			writer.Put("txPreview_ByHash"+tx.Bloom.HashStr, buffer)

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

			writer.Put("txKeys:"+tx.Bloom.HashStr, keysArrayMarshal)

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

				writer.Put("addrTx:"+keyStr+":"+strconv.FormatUint(count, 10), tx.Bloom.Hash)
				writer.Put("addrTxsCount:"+keyStr, []byte(strconv.FormatUint(count+1, 10)))
			}

		}
	}

	return
}

func saveBlockCompleteInfo(writer store_db_interface.StoreDBTransactionInterface, blkComplete *block_complete.BlockComplete, transactionsCount uint64, localTransactionChanges []*blockchain_types.BlockchainTransactionUpdate) (err error) {

	var fees uint64
	if fees, err = blkComplete.ComputeFees(); err != nil {
		return
	}

	var blockInfoMarshal []byte
	if blockInfoMarshal, err = json.Marshal(&info.BlockInfo{
		Hash:       blkComplete.Block.Bloom.Hash,
		KernelHash: blkComplete.Block.Bloom.KernelHash,
		Timestamp:  blkComplete.Block.Timestamp,
		Size:       blkComplete.BloomBlkComplete.Size,
		TXs:        uint64(len(blkComplete.Txs)),
		Fees:       fees,
		Forger:     blkComplete.Block.Forger,
	}); err != nil {
		return
	}

	writer.Put("blockInfo_ByHash"+string(blkComplete.Block.Bloom.Hash), blockInfoMarshal)

	for i, tx := range blkComplete.Txs {

	}

	return
}
