package mempool

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sync"
)

type MemPoolTx struct {
	Tx         *transaction.Transaction
	FeePerByte uint64
}

type MemPoolType struct {
	txs          sync.Map
	sortedByFees []helpers.Hash

	sync.Mutex
}

var MemPool MemPoolType

func (mempool *MemPoolType) AddTxToMemPool(tx *transaction.Transaction) (result bool) {

	mempool.Lock()
	defer mempool.Unlock()

	hash := tx.ComputeHash()
	if _, found := mempool.txs.Load(hash); found {
		return false
	}

	object := MemPoolTx{
		Tx: tx,
	}

	return true
}

func InitMemPool() (err error) {

	gui.Info("MemPool init...")

	return
}
