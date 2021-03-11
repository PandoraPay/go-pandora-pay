package mempool

import (
	"bytes"
	"fmt"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/cryptography"
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

func (mempool *MemPool) GetTxsListKeyValue() []*memPoolOutput {

	list := make([]*memPoolOutput, 0)

	mempool.txs.Range(func(key, value interface{}) bool {
		hash := key.(cryptography.Hash)
		tx := value.(*memPoolTx)
		list = append(list, &memPoolOutput{
			hash:    hash,
			hashStr: string(hash[:]),
			tx:      tx,
		})
		return true
	})

	return list
}

func (mempool *MemPool) GetTxsListKeyValueFilter(filter map[string]bool) []*memPoolOutput {

	list := make([]*memPoolOutput, 0)

	mempool.txs.Range(func(key, value interface{}) bool {
		hash := key.(cryptography.Hash)
		if !filter[string(hash[:])] {
			tx := value.(*memPoolTx)
			list = append(list, &memPoolOutput{
				hash:    hash,
				hashStr: string(hash[:]),
				tx:      tx,
			})
		}
		return true
	})

	return list
}

func (mempool *MemPool) GetNonce(publicKeyHash [20]byte) (result bool, nonce uint64) {

	txs := mempool.GetTxsList()
	for _, tx := range txs {
		if tx.TxType == transaction_type.TxSimple {
			base := tx.TxBase.(*transaction_simple.TransactionSimple)
			txPublicKeyHash := base.Vin[0].GetPublicKeyHash()
			if bytes.Equal(txPublicKeyHash[:], publicKeyHash[:]) {
				result = true
				if nonce <= base.Nonce {
					nonce = base.Nonce + 1
				}
			}
		}
	}

	return
}

func (mempool *MemPool) GetTransactions(blockHeight uint64, chainHash cryptography.Hash) []*transaction.Transaction {

	out := make([]*transaction.Transaction, 0)

	mempool.result.RLock()
	for _, mempoolTx := range mempool.result.txs {
		if bytes.Equal(mempool.result.chainHash[:], chainHash[:]) {
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
		out.tx.RLock()
		gui.Log(fmt.Sprintf("%20s %7d B %5d %32s", time.Unix(out.tx.added, 0).UTC().Format(time.RFC3339), len(out.tx.tx.Serialize()), out.tx.chainHeight, out.hash.String()))
		out.tx.RUnlock()
	}
	gui.Log("")

}
