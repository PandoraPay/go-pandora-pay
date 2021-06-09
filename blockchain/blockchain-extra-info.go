package blockchain

import (
	"encoding/json"
	"errors"
	block_complete "pandora-pay/blockchain/blocks/block-complete"
	"pandora-pay/blockchain/info"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/tokens/token"
	store_db_interface "pandora-pay/store/store-db/store-db-interface"
	"strconv"
)

func removeBlockInfo(writer store_db_interface.StoreDBTransactionInterface, hash []byte) (err error) {
	return writer.Delete("blockInfo_ByHash" + string(hash))
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
	}

	return
}

func saveTokensInfo(toks *tokens.Tokens) (err error) {

	for k, v := range toks.Committed {

		if v.Stored == "del" {
			err = toks.Tx.DeleteForcefully("tokenInfo_ByHash" + k)
		} else if v.Stored == "update" {

			tok := v.Element.(*token.Token)
			tokInfo := &info.TokenInfo{
				Name:             tok.Name,
				Ticker:           tok.Ticker,
				DecimalSeparator: tok.DecimalSeparator,
				Description:      tok.Description,
			}
			var data []byte
			if data, err = json.Marshal(tokInfo); err != nil {
				return
			}

			err = toks.Tx.Put("tokenInfo_ByHash"+k, data)
		}

		if err != nil {
			return
		}
	}

	return
}

func saveBlockCompleteInfo(writer store_db_interface.StoreDBTransactionInterface, blkComplete *block_complete.BlockComplete, transactionsCount uint64) (err error) {

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
	}

	return
}
