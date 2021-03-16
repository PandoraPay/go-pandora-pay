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

func (mempool *Mempool) GetTxsList() []*mempoolTx {

	mempool.txs.RLock()
	transactions := make([]*mempoolTx, len(mempool.txs.txsList))
	copy(transactions, mempool.txs.txsList)
	mempool.txs.RUnlock()

	return transactions
}

func (mempool *Mempool) GetNonce(publicKeyHash []byte) (result bool, nonce uint64) {

	txs := mempool.GetTxsList()
	for _, tx := range txs {
		if tx.Tx.TxType == transaction_type.TxSimple {
			base := tx.Tx.TxBase.(*transaction_simple.TransactionSimple)
			if bytes.Equal(base.Vin[0].Bloom.PublicKeyHash, publicKeyHash) {
				result = true
				if nonce <= base.Nonce {
					nonce = base.Nonce + 1
				}
			}
		}
	}

	return
}

func (mempool *Mempool) GetNextTransactionsToInclude(blockHeight uint64, chainHash []byte) (out []*transaction.Transaction) {

	mempool.result.RLock()
	if bytes.Equal(mempool.result.chainHash, chainHash) {
		out = make([]*transaction.Transaction, len(mempool.result.txs))
		copy(out, mempool.result.txs)
	} else {
		out = []*transaction.Transaction{}
	}
	mempool.result.RUnlock()
	return
}

func (mempool *Mempool) print() {

	transactions := mempool.GetTxsList()

	if len(transactions) == 0 {
		return
	}

	gui.Log("")
	for _, out := range transactions {
		gui.Log(fmt.Sprintf("%20s %7d B %5d %15s", time.Unix(out.Added, 0).UTC().Format(time.RFC3339), out.Tx.Bloom.Size, out.ChainHeight, hex.EncodeToString(out.Tx.Bloom.Hash[0:15])))
	}
	gui.Log("")

}
