package blockchain

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"pandora-pay/blockchain/blockchain_types"
	"pandora-pay/blockchain/blocks/block_complete"
	"pandora-pay/blockchain/data_storage/assets"
	"pandora-pay/blockchain/data_storage/assets/asset"
	"pandora-pay/blockchain/info"
	"pandora-pay/helpers/generics"
	"pandora-pay/store/store_db/store_db_interface"
	"strconv"
)

func removeBlockCompleteInfo(writer store_db_interface.StoreDBTransactionInterface, hash []byte, txHashes [][]byte, localTransactionChanges []*blockchain_types.BlockchainTransactionUpdate) (err error) {

	writer.Delete("blockInfo_ByHash" + string(hash))

	for i, txHash := range txHashes {

		data := writer.Get("txKeys:" + string(txHash))
		if data == nil {
			return errors.New("TxKeys is missing")
		}
		keys := make([][]byte, 0)
		if err = msgpack.Unmarshal(data, &keys); err != nil {
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
			writer.Delete("addrTx:" + string(key) + ":" + strconv.FormatUint(count, 10))
			if count == 0 {
				writer.Delete("addrTxsCount:" + string(key))
			} else {
				writer.Put("addrTxsCount:"+string(key), []byte(strconv.FormatUint(count, 10)))
			}
		}

		writer.Delete("txKeys:" + string(txHash))
	}

	return
}

func removeUnusedTransactions(writer store_db_interface.StoreDBTransactionInterface, starting, count uint64) {

	for i := starting; i < count; i++ {
		writer.Delete("txHash_ByHeight" + strconv.FormatUint(i, 10))
	}
}

func removeTxsInfo(writer store_db_interface.StoreDBTransactionInterface, removedTxHashes map[string][]byte) {

	for txHash := range removedTxHashes {
		writer.Delete("txInfo_ByHash" + txHash)
		writer.Delete("txPreview_ByHash" + txHash)
	}
}

func saveAssetsInfo(asts *assets.Assets) (err error) {

	for k, v := range asts.Committed {

		if v.Stored == "del" {
			asts.Tx.Delete("assetInfo_ByHash:" + k)
		} else if v.Stored == "update" {

			ast := v.Element.(*asset.Asset)
			astInfo := &info.AssetInfo{
				ast.Version,
				ast.Name,
				ast.Ticker,
				ast.DecimalSeparator,
				ast.Description[:generics.Min(100, len(ast.Description))],
				[]byte(k),
			}
			var data []byte
			if data, err = msgpack.Marshal(astInfo); err != nil {
				return
			}

			asts.Tx.Put("assetInfo_ByHash:"+k, data)
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
	if blockInfoMarshal, err = msgpack.Marshal(&info.BlockInfo{
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

		height := transactionsCount + uint64(i)
		indexStr := strconv.FormatUint(height, 10)
		writer.Put("txHash_ByHeight"+indexStr, tx.Bloom.Hash)

		var buffer []byte
		if buffer, err = msgpack.Marshal(&info.TxInfo{
			height,
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
		if buffer, err = msgpack.Marshal(txPreview); err != nil {
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
		if keysArrayMarshal, err = msgpack.Marshal(keysArray); err != nil {
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

	return
}
