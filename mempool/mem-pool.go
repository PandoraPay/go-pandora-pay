package mempool

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/fees"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sync"
)

type MemPoolTx struct {
	Tx         *transaction.Transaction
	Height     uint64
	FeePerByte uint64
	FeeToken   []byte
}

type MemPool struct {
	txs          sync.Map
	sortedByFees []helpers.Hash

	sync.Mutex
}

func (mempool *MemPool) AddTxToMemPool(tx *transaction.Transaction, height uint64) (result bool) {

	var err error

	mempool.Lock()
	defer mempool.Unlock()

	hash := tx.ComputeHash()
	if _, found := mempool.txs.Load(hash); found {
		return
	}

	var minerFees map[string]uint64
	if minerFees, err = tx.ComputeFees(); err != nil {
		return false
	}
	size := uint64(len(tx.Serialize(true)))
	var selectedFeeToken *string
	for token := range fees.FEES_PER_BYTE {
		if minerFees[token] != 0 {
			feePerByte := minerFees[token] / size
			if feePerByte >= fees.FEES_PER_BYTE[token] {
				selectedFeeToken = &token
				break
			}
		}
	}
	if selectedFeeToken == nil {
		return
	}

	object := MemPoolTx{
		Tx:         tx,
		Height:     height,
		FeePerByte: minerFees[*selectedFeeToken] / size,
		FeeToken:   []byte(*selectedFeeToken),
	}

	mempool.txs.Store(hash, object)

	return true
}

func (mempool *MemPool) Exists(txId helpers.Hash) bool {
	if _, exists := mempool.txs.Load(txId); exists {
		return true
	}
	return false
}

func (mempool *MemPool) Delete(txId helpers.Hash) (tx *transaction.Transaction) {

	var exists bool
	var objInterface interface{}
	if objInterface, exists = mempool.txs.Load(txId); exists {
		return nil
	}

	object := objInterface.(*MemPoolTx)
	tx = object.Tx
	mempool.txs.Delete(txId)

	return
}

func InitMemPool() (mempool *MemPool, err error) {

	gui.Log("MemPool init...")

	mempool = new(MemPool)

	return
}
