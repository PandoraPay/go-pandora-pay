package mempool

import (
	"fmt"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config/fees"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sync"
	"time"
)

type MemPoolTx struct {
	tx         *transaction.Transaction
	added      int64
	mine       bool
	height     uint64
	feePerByte uint64
	feeToken   []byte
}

type MemPool struct {
	txs          sync.Map
	sortedByFees []helpers.Hash

	sync.RWMutex
}

func (mempool *MemPool) AddTxToMemPool(tx *transaction.Transaction, height uint64, mine bool) (result bool) {

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
		tx:         tx,
		added:      time.Now().Unix(),
		height:     height,
		feePerByte: minerFees[*selectedFeeToken] / size,
		feeToken:   []byte(*selectedFeeToken),
		mine:       mine,
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
	tx = object.tx
	mempool.txs.Delete(txId)

	return
}

func (mempool *MemPool) Print() {

	type Output struct {
		hash helpers.Hash
		tx   *MemPoolTx
	}
	var list []*Output

	mempool.txs.Range(func(key, value interface{}) bool {
		hash := key.(helpers.Hash)
		tx := value.(*MemPoolTx)
		list = append(list, &Output{
			hash,
			tx,
		})

		return true
	})

	gui.Log("")
	gui.Log(fmt.Sprintf("TX mempool: %d", len(list)))
	for _, out := range list {
		gui.Log(fmt.Sprintf("%20s %7d B %5d %32s", time.Unix(out.tx.added, 0).UTC().Format(time.RFC3339), len(out.tx.tx.Serialize(true)), out.tx.height, out.hash))
	}
	gui.Log("")

}

func InitMemPool() (mempool *MemPool, err error) {

	gui.Log("MemPool init...")

	mempool = new(MemPool)

	go func() {
		for {
			time.Sleep(60 * time.Second)
			mempool.Print()
		}
	}()

	return
}
