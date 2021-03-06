package mempool

import (
	"fmt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config/fees"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type memPoolTx struct {
	tx         *transaction.Transaction
	added      int64
	mine       bool
	height     uint64
	feePerByte uint64
	feeToken   []byte
}

type memPoolUpdateTask struct {
	hash   helpers.Hash
	height uint64
	accs   *accounts.Accounts
	toks   *tokens.Tokens

	sync.RWMutex
}

type memPoolOutput struct {
	hash helpers.Hash
	tx   *memPoolTx
}

type MemPool struct {
	txs sync.Map

	updateTask      memPoolUpdateTask
	updateTaskReset int32

	mutex sync.Mutex `json:"-"`
}

func (mempool *MemPool) AddTxToMemPoolSilent(tx *transaction.Transaction, height uint64, mine bool) (result bool, err error) {
	defer func() {
		if err2 := recover(); err2 != nil {
			err = helpers.ConvertRecoverError(err2)
		}
	}()
	result = mempool.AddTxToMemPool(tx, height, mine)
	return
}

func (mempool *MemPool) AddTxToMemPool(tx *transaction.Transaction, height uint64, mine bool) bool {

	//making sure that the transaction is not inserted twice
	mempool.mutex.Lock()
	defer mempool.mutex.Unlock()

	hash := tx.ComputeHash()
	if _, found := mempool.txs.Load(hash); found {
		return false
	}

	minerFees := tx.ComputeFees()

	size := uint64(len(tx.Serialize()))
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
		panic("Transaction fee was not accepted")
	}

	object := memPoolTx{
		tx:         tx,
		added:      time.Now().Unix(),
		height:     height,
		feePerByte: minerFees[*selectedFeeToken] / size,
		feeToken:   []byte(*selectedFeeToken),
		mine:       mine,
	}

	mempool.txs.Store(hash, &object)

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

	object := objInterface.(*memPoolTx)
	tx = object.tx
	mempool.txs.Delete(txId)

	return
}

func (mempool *MemPool) getTxsList() []*memPoolOutput {

	list := []*memPoolOutput{}

	mempool.txs.Range(func(key, value interface{}) bool {
		hash := key.(helpers.Hash)
		tx := value.(*memPoolTx)
		list = append(list, &memPoolOutput{
			hash,
			tx,
		})
		return true
	})

	return list
}

func (mempool *MemPool) Print() {

	list := mempool.getTxsList()

	gui.Log("")
	gui.Log(fmt.Sprintf("TX mempool: %d", len(list)))
	for _, out := range list {
		gui.Log(fmt.Sprintf("%20s %7d B %5d %32s", time.Unix(out.tx.added, 0).UTC().Format(time.RFC3339), len(out.tx.tx.Serialize()), out.tx.height, out.hash))
	}
	gui.Log("")

}

func (mempool *MemPool) UpdateChanges(hash helpers.Hash, height uint64, accs *accounts.Accounts, toks *tokens.Tokens) {
	mempool.updateTask.Lock()
	defer mempool.updateTask.Unlock()
	copy(mempool.updateTask.hash[:], hash[:])
	mempool.updateTask.height = height
	mempool.updateTask.accs = accs
	mempool.updateTask.toks = toks
	atomic.StoreInt32(&mempool.updateTaskReset, 1)
}

func (mempool *MemPool) Refresh() {

	updateTask := memPoolUpdateTask{}
	hasWorkToDo := false
	var list []*memPoolOutput
	listIndex := -1
	for {

		value := atomic.LoadInt32(&mempool.updateTaskReset)
		if value == 1 {
			mempool.updateTask.RLock()
			copy(updateTask.hash[:], mempool.updateTask.hash[:])
			updateTask.height = mempool.updateTask.height
			updateTask.accs = mempool.updateTask.accs
			updateTask.toks = mempool.updateTask.toks
			hasWorkToDo = true
			listIndex = -1
			atomic.StoreInt32(&mempool.updateTaskReset, 2)
			mempool.updateTask.RUnlock()
		}

		if hasWorkToDo {

			if listIndex == -1 {

				list = mempool.getTxsList()
				if len(list) > 0 {
					sort.Slice(list, func(i, j int) bool {

						if list[i].tx.feePerByte == list[j].tx.feePerByte {

							if list[i].tx.tx.TxType == transaction_type.TxSimple && list[j].tx.tx.TxType == transaction_type.TxSimple {
								return list[i].tx.tx.TxBase.(*transaction_simple.TransactionSimple).Nonce < list[j].tx.tx.TxBase.(*transaction_simple.TransactionSimple).Nonce
							}

						}

						return list[i].tx.feePerByte < list[j].tx.feePerByte
					})
				}
				listIndex = 0

			} else {

				if listIndex == len(list) {
					hasWorkToDo = false
					continue
				} else {
					list[listIndex].tx.height = updateTask.height
					continue
				}

			}

		} else {
			time.Sleep(100 * time.Millisecond)
		}

	}
}

func InitMemPool() (mempool *MemPool) {

	gui.Log("MemPool init...")

	mempool = &MemPool{}

	go func() {
		for {
			time.Sleep(60 * time.Second)
			mempool.Print()
		}
	}()

	go mempool.Refresh()

	return
}
