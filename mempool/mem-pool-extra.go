package mempool

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/gui"
	"time"
)

func (mempool *MemPool) GetTxsList() []*memPoolTx {

	mempool.lockWritingTxs.RLock()
	transactions := make([]*memPoolTx, len(mempool.txsList))
	copy(transactions, mempool.txsList)
	mempool.lockWritingTxs.RUnlock()

	return transactions
}

func (mempool *MemPool) GetTxsListKeyValueFilter(filter map[string]bool) []*memPoolTx {

	list := []*memPoolTx{}

	mempool.txs.Range(func(key, value interface{}) bool {
		hash := key.(string)
		if !filter[hash] {
			tx := value.(*memPoolTx)
			list = append(list, tx)
		}
		return true
	})

	return list
}

func (mempool *MemPool) GetNonce(publicKeyHash []byte) (result bool, nonce uint64) {

	txs := mempool.GetTxsList()
	for _, tx := range txs {
		if tx.Tx.TxType == transaction_type.TxSimple {
			base := tx.Tx.TxBase.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin[0].GetPublicKeyHash(), publicKeyHash) {
				result = true
				if nonce <= base.Nonce {
					nonce = base.Nonce + 1
				}
			}
		}
	}

	return
}

func (mempool *MemPool) GetNextTransactionsToInclude(blockHeight uint64, chainHash []byte) (out []*transaction.Transaction) {

	mempool.result.RLock()
	if bytes.Equal(mempool.result.chainHash, chainHash) {
		out = make([]*transaction.Transaction, len(mempool.result.txs))
		for i, mempoolTx := range mempool.result.txs {
			out[i] = mempoolTx.Tx
		}
	} else {
		out = []*transaction.Transaction{}
	}
	mempool.result.RUnlock()
	return
}

func (mempool *MemPool) print() {

	transactions := mempool.GetTxsList()

	if len(transactions) == 0 {
		return
	}

	gui.Log("")
	for _, out := range transactions {
		gui.Log(fmt.Sprintf("%20s %7d B %5d %15s", time.Unix(out.Added, 0).UTC().Format(time.RFC3339), out.Size, out.ChainHeight, hex.EncodeToString(out.Hash[0:15])))
	}
	gui.Log("")

}
