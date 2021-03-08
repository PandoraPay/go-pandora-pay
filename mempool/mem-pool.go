package mempool

import (
	"bytes"
	"fmt"
	"pandora-pay/blockchain/accounts"
	"pandora-pay/blockchain/tokens"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/config/fees"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type memPoolTx struct {
	tx          *transaction.Transaction
	added       int64
	mine        bool
	chainHeight uint64
	chainHash   helpers.Hash
	feePerByte  uint64
	feeToken    []byte
}

type memPoolUpdateTask struct {
	chainHash   helpers.Hash
	chainHeight uint64
	accs        *accounts.Accounts
	toks        *tokens.Tokens

	sync.RWMutex
}

type memPoolOutput struct {
	hash helpers.Hash
	tx   *memPoolTx
}

type MemPool struct {
	txs      sync.Map
	txsCount uint64

	updateTask      memPoolUpdateTask
	updateTaskReset int32

	sync.RWMutex `json:"-"`
}

func (mempool *MemPool) AddTxToMemPoolSilent(tx *transaction.Transaction, height uint64, mine bool) (result bool, err error) {
	defer func() {
		err = helpers.ConvertRecoverError(recover())
	}()
	result = mempool.AddTxToMemPool(tx, height, mine)
	return
}

func (mempool *MemPool) AddTxToMemPool(tx *transaction.Transaction, height uint64, mine bool) bool {

	//making sure that the transaction is not inserted twice
	mempool.Lock()
	defer mempool.Unlock()

	hash := tx.ComputeHash()
	if _, found := mempool.txs.Load(hash); found {
		return false
	}

	mempool.txsCount += 1
	gui.Info2Update("mempool", strconv.FormatUint(mempool.txsCount, 10))

	minerFees := tx.ComputeFees()

	size := uint64(len(tx.Serialize()))
	var selectedFeeToken *string
	var selectedFee uint64

	for token := range fees.FEES_PER_BYTE {
		if minerFees[token] != 0 {
			feePerByte := minerFees[token] / size
			if feePerByte >= fees.FEES_PER_BYTE[token] {
				selectedFeeToken = &token
				selectedFee = minerFees[*selectedFeeToken]
				break
			}
		}
	}

	//if it is mine and no fee was paid, let's fake a fee
	if mine && selectedFeeToken == nil {
		nativeFee := config.NATIVE_TOKEN_STRING
		selectedFeeToken = &nativeFee
		selectedFee = fees.FEES_PER_BYTE[config.NATIVE_TOKEN_STRING]
	}

	if selectedFeeToken == nil {
		panic("Transaction fee was not accepted")
	}

	object := memPoolTx{
		tx:          tx,
		added:       time.Now().Unix(),
		chainHeight: height,
		feePerByte:  selectedFee / size,
		feeToken:    []byte(*selectedFeeToken),
		mine:        mine,
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
	if _, exists = mempool.txs.Load(txId); exists == false {
		return
	}

	mempool.Lock()
	defer mempool.Unlock()

	mempool.txs.Delete(txId)

	mempool.txsCount -= 1
	gui.Info2Update("mempool", strconv.FormatUint(mempool.txsCount, 10))

	return
}

func (mempool *MemPool) GetTxsList() []*memPoolTx {

	list := make([]*memPoolTx, 0)

	mempool.txs.Range(func(key, value interface{}) bool {
		tx := value.(*memPoolTx)
		list = append(list, tx)
		return true
	})

	return list
}

func (mempool *MemPool) GetTxsListKeyValue() []*memPoolOutput {

	list := make([]*memPoolOutput, 0)

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

	mempool.RLock()
	defer mempool.RUnlock()

	if mempool.txsCount == 0 {
		return
	}

	list := mempool.GetTxsListKeyValue()

	gui.Log("")
	for _, out := range list {
		gui.Log(fmt.Sprintf("%20s %7d B %5d %32s", time.Unix(out.tx.added, 0).UTC().Format(time.RFC3339), len(out.tx.tx.Serialize()), out.tx.chainHeight, out.hash.String()))
	}
	gui.Log("")

}

func (mempool *MemPool) UpdateChanges(hash helpers.Hash, height uint64, accs *accounts.Accounts, toks *tokens.Tokens) {

	mempool.updateTask.Lock()
	defer mempool.updateTask.Unlock()

	mempool.updateTask.chainHash = hash
	mempool.updateTask.chainHeight = height
	mempool.updateTask.accs = accs
	mempool.updateTask.toks = toks

	atomic.StoreInt32(&mempool.updateTaskReset, 1)
}

func (mempool *MemPool) Refresh() {

	updateTask := memPoolUpdateTask{}
	hasWorkToDo := false
	var list []*memPoolTx

	listIndex := -1
	for {

		value := atomic.LoadInt32(&mempool.updateTaskReset)
		if value == 1 {

			mempool.updateTask.RLock()
			updateTask.chainHash = mempool.updateTask.chainHash
			updateTask.chainHeight = mempool.updateTask.chainHeight
			updateTask.accs = mempool.updateTask.accs
			updateTask.toks = mempool.updateTask.toks
			hasWorkToDo = true
			listIndex = -1
			atomic.StoreInt32(&mempool.updateTaskReset, 2)
			mempool.updateTask.RUnlock()
		}

		if hasWorkToDo {

			if listIndex == -1 {

				list = mempool.GetTxsList()

				if len(list) > 0 {
					sort.Slice(list, func(i, j int) bool {

						if list[i].feePerByte == list[j].feePerByte && list[i].tx.TxType == transaction_type.TxSimple && list[j].tx.TxType == transaction_type.TxSimple {
							return list[i].tx.TxBase.(*transaction_simple.TransactionSimple).Nonce < list[j].tx.TxBase.(*transaction_simple.TransactionSimple).Nonce
						}

						return list[i].feePerByte < list[j].feePerByte
					})
				}
				listIndex = 0

			} else {

				if listIndex == len(list) {
					hasWorkToDo = false
					continue
				} else {

					func() {
						defer func() {
							err := helpers.ConvertRecoverError(recover())
							if err != nil {
								updateTask.accs.Rollback()
								updateTask.toks.Rollback()
							} else {
								list[listIndex].chainHeight = updateTask.chainHeight
								list[listIndex].chainHash = updateTask.chainHash
							}
							listIndex += 1
						}()

						list[listIndex].tx.IncludeTransaction(updateTask.chainHeight, updateTask.accs, updateTask.toks)
					}()

					continue
				}

			}

		} else {
			time.Sleep(100 * time.Millisecond)
		}

	}
}

func (mempool *MemPool) GetTransactions(blockHeight uint64, chainHash helpers.Hash) []*transaction.Transaction {

	out := make([]*transaction.Transaction, 0)

	list := mempool.GetTxsList()
	for _, mempoolTx := range list {
		if mempoolTx.chainHeight == blockHeight && bytes.Equal(mempoolTx.chainHash[:], chainHash[:]) {
			out = append(out, mempoolTx.tx)
		}
	}

	return out
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
