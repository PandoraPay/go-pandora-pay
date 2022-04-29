package mempool

import (
	"bytes"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"sort"
)

type ContinueProcessingType byte

const (
	CONTINUE_PROCESSING_ERROR ContinueProcessingType = iota
	CONTINUE_PROCESSING_NO_ERROR_RESET
	CONTINUE_PROCESSING_NO_ERROR
)

func (mempool *Mempool) ExistsTxSimpleVersion(publicKeyHash []byte, version transaction_simple.ScriptType) bool {

	txs := mempool.Txs.GetTxsList()
	if txs == nil {
		return false
	}

	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if base.TxScript == version {
				for i := range base.Vin {
					if bytes.Equal(base.Bloom.VinPublicKeyHashes[i], publicKeyHash) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (mempool *Mempool) CountInputTxs(publicKeyHash []byte) uint64 {

	txs := mempool.Txs.GetTxsList()

	count := uint64(0)
	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			for i := range base.Vin {
				if bytes.Equal(base.Bloom.VinPublicKeyHashes[i], publicKeyHash) {
					count++
				}
			}
		}
	}

	return count
}

func (mempool *Mempool) GetNonce(publicKeyHash []byte, nonce uint64) uint64 {

	txs := mempool.Txs.GetTxsList()

	nonces := make(map[uint64]bool)
	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Bloom.VinPublicKeyHashes[0], publicKeyHash) {
				nonces[base.Nonce] = true
			}
		}
	}

	for {
		if nonces[nonce] {
			nonce += 1
		} else {
			break
		}
	}

	return nonce

}

func (mempool *Mempool) GetNextTransactionsToInclude(chainHash []byte) (out []*transaction.Transaction, outChainHash []byte) {

	res := mempool.result.Load()
	if res != nil {

		if chainHash == nil || bytes.Equal(res.chainHash, chainHash) {
			txs := res.txs.Load()
			finalTxs := make([]*transaction.Transaction, len(txs))
			for i, tx := range txs {
				finalTxs[i] = tx.Tx
			}
			return finalTxs, res.chainHash
		}
	}

	return []*transaction.Transaction{}, nil
}

func sortTxs(txList []*mempoolTx) {
	sort.Slice(txList, func(i, j int) bool {

		if txList[i].FeePerByte == txList[j].FeePerByte && txList[i].Tx.Version == transaction_type.TX_SIMPLE && txList[j].Tx.Version == transaction_type.TX_SIMPLE {
			return txList[i].Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce < txList[j].Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce
		}

		return txList[i].FeePerByte < txList[j].FeePerByte
	})
}
