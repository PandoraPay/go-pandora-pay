package mempool

import (
	"bytes"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/helpers"
	"sort"
)

func (mempool *Mempool) GetBalance(publicKeyHash []byte, balance uint64, token []byte) (out uint64, err error) {

	out = balance
	txs := mempool.Txs.GetTxsList()

	for _, tx := range txs {
		if tx.Tx.TxType == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			for _, vin := range base.Vin {
				if bytes.Equal(vin.Bloom.PublicKeyHash, publicKeyHash) && bytes.Equal(vin.Token, token) {
					if err = helpers.SafeUint64Sub(&out, vin.Amount); err != nil {
						return
					}
				}
			}

			for _, vout := range base.Vout {
				if bytes.Equal(vout.PublicKeyHash, publicKeyHash) && bytes.Equal(vout.Token, token) {
					if err = helpers.SafeUint64Add(&out, vout.Amount); err != nil {
						return
					}
				}
			}

		}
	}

	return
}

func (mempool *Mempool) ExistsTxSimpleVersion(publicKeyHash []byte, version transaction_simple.ScriptType) bool {

	txs := mempool.Txs.GetTxsList()
	for _, tx := range txs {
		if tx.Tx.TxType == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin[0].Bloom.PublicKeyHash, publicKeyHash) && base.TxScript == version {
				return true
			}
		}
	}
	return false
}

func (mempool *Mempool) CountInputTxs(publicKeyHash []byte) uint64 {

	txs := mempool.Txs.GetTxsList()

	count := uint64(0)
	for _, tx := range txs {
		if tx.Tx.TxType == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin[0].Bloom.PublicKeyHash, publicKeyHash) {
				count++
			}
		}
	}

	return count
}

func (mempool *Mempool) GetNonce(publicKeyHash []byte, nonce uint64) uint64 {

	txs := mempool.Txs.GetTxsList()

	nonces := make(map[uint64]bool)
	for _, tx := range txs {
		if tx.Tx.TxType == transaction_type.TX_SIMPLE {
			base := tx.Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin[0].Bloom.PublicKeyHash, publicKeyHash) {
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

func (mempool *Mempool) GetNextTransactionsToInclude(blockHeight uint64, chainHash []byte) (out []*transaction.Transaction) {

	result := mempool.result.Load()
	if result != nil {

		res := result.(*MempoolResult)

		if bytes.Equal(res.chainHash, chainHash) {
			txs := res.txs.Load().([]*mempoolTx)
			finaltxs := make([]*transaction.Transaction, len(txs))
			for i, tx := range txs {
				finaltxs[i] = tx.Tx
			}
			return finaltxs
		}
	}

	return []*transaction.Transaction{}
}

func sortTxs(txList []*mempoolTx) {
	sort.Slice(txList, func(i, j int) bool {

		if txList[i].FeePerByte == txList[j].FeePerByte && txList[i].Tx.TxType == transaction_type.TX_SIMPLE && txList[j].Tx.TxType == transaction_type.TX_SIMPLE {
			return txList[i].Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce < txList[j].Tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple).Nonce
		}

		return txList[i].FeePerByte < txList[j].FeePerByte
	})
}
