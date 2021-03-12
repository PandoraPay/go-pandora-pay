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

func (mempool *MemPool) GetTxsList() []*transaction.Transaction {

	list := make([]*transaction.Transaction, 0)

	mempool.txs.Range(func(key, value interface{}) bool {
		tx := value.(*memPoolTx)
		list = append(list, tx.tx)
		return true
	})

	return list
}

func (mempool *MemPool) GetTxsListKeyValue() []*memPoolTx {

	list := make([]*memPoolTx, 0)

	mempool.txs.Range(func(key, value interface{}) bool {
		tx := value.(*memPoolTx)
		list = append(list, tx)
		return true
	})

	return list
}

func (mempool *MemPool) GetTxsListKeyValueFilter(filter map[string]bool) []*memPoolTx {

	list := make([]*memPoolTx, 0)

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
		if tx.TxType == transaction_type.TxSimple {
			base := tx.TxBase.(*transaction_simple.TransactionSimple)
			txPublicKeyHash := base.Vin[0].GetPublicKeyHash()
			if bytes.Equal(txPublicKeyHash, publicKeyHash) {
				result = true
				if nonce <= base.Nonce {
					nonce = base.Nonce + 1
				}
			}
		}
	}

	return
}

func (mempool *MemPool) GetTransactions(blockHeight uint64, chainHash []byte) []*transaction.Transaction {

	out := make([]*transaction.Transaction, 0)

	mempool.result.RLock()
	if bytes.Equal(mempool.result.chainHash, chainHash) {
		for _, mempoolTx := range mempool.result.txs {
			out = append(out, mempoolTx.tx)
		}
	}
	mempool.result.RUnlock()

	return out
}

func (mempool *MemPool) print() {

	mempool.lockWritingTxs.RLock()
	txsCount := mempool.txsCount
	mempool.lockWritingTxs.RUnlock()

	if txsCount == 0 {
		return
	}

	list := mempool.GetTxsListKeyValue()

	gui.Log("")
	for _, out := range list {
		out.RLock()
		gui.Log(fmt.Sprintf("%20s %7d B %5d %32s", time.Unix(out.added, 0).UTC().Format(time.RFC3339), len(out.tx.Serialize()), out.chainHeight, hex.EncodeToString(out.hash)))
		out.RUnlock()
	}
	gui.Log("")

}
