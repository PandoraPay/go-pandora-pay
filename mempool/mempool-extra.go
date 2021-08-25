package mempool

import (
	"bytes"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/helpers"
	"sort"
)

type ContinueProcessingType byte

const (
	CONTINUE_PROCESSING_ERROR ContinueProcessingType = iota
	CONTINUE_PROCESSING_NO_ERROR_RESET
	CONTINUE_PROCESSING_NO_ERROR
)

func (mempool *Mempool) GetBalance(publicKey []byte, balance uint64, token []byte) (uint64, error) {

	out := balance
	txs := mempool.Txs.GetTxsFromMap()

	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			for _, vin := range base.Vin {
				if bytes.Equal(vin.PublicKey, publicKey) && bytes.Equal(vin.Token, token) {
					if err := helpers.SafeUint64Sub(&out, vin.Amount); err != nil {
						return 0, err
					}
				}
			}

			for _, vout := range base.Vout {
				if bytes.Equal(vout.PublicKey, publicKey) && bytes.Equal(vout.Token, token) {
					if err := helpers.SafeUint64Add(&out, vout.Amount); err != nil {
						return 0, err
					}
				}
			}

		}
	}

	return out, nil
}

func (mempool *Mempool) ExistsTxSimpleVersion(publicKey []byte, version transaction_simple.ScriptType) bool {

	txs := mempool.Txs.GetTxsFromMap()
	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin[0].PublicKey, publicKey) && base.TxScript == version {
				return true
			}
		}
	}
	return false
}

func (mempool *Mempool) CountInputTxs(publicKey []byte) uint64 {

	txs := mempool.Txs.GetTxsFromMap()

	count := uint64(0)
	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin[0].PublicKey, publicKey) {
				count++
			}
		}
	}

	return count
}

func (mempool *Mempool) GetNonce(publicKey []byte, nonce uint64) uint64 {

	txs := mempool.Txs.GetTxsFromMap()

	nonces := make(map[uint64]bool)
	for _, tx := range txs {
		if tx.Tx.Version == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin[0].PublicKey, publicKey) {
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

	result := mempool.result.Load()
	if result != nil {

		res := result.(*MempoolResult)

		if chainHash == nil || bytes.Equal(res.chainHash, chainHash) {
			txs := res.txs.Load().([]*mempoolTx)
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
