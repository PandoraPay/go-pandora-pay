package mempool

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/fees"
	"pandora-pay/gui"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type mempoolTx struct {
	Tx          *transaction.Transaction
	Added       int64
	Mine        bool
	FeePerByte  uint64
	FeeToken    []byte //20 byte
	ChainHeight uint64
}

type mempoolResult struct {
	txs          []*transaction.Transaction
	totalSize    uint64
	chainHash    []byte //32
	chainHeight  uint64
	sync.RWMutex `json:"-"`
}

type mempoolTxs struct {
	txsCount     int64        //use atomic
	txsInserted  int64        //use atomic
	txsList      atomic.Value // []*mempoolTx
	txsListMutex sync.Mutex   // for writing
}

type Mempool struct {
	txs     *mempoolTxs
	result  *mempoolResult
	newWork chan *mempoolWork
}

func (mempool *Mempool) AddTxToMemPool(tx *transaction.Transaction, height uint64, mine bool) (out bool, err error) {
	return mempool.AddTxsToMemPool([]*transaction.Transaction{tx}, height, mine)
}

func (mempool *Mempool) AddTxsToMemPool(txs []*transaction.Transaction, height uint64, mine bool) (out bool, err error) {

	finalTxs := []*mempoolTx{}

	for _, tx := range txs {
		if err = tx.VerifyBloomAll(); err != nil {
			return
		}

		var minerFees map[string]uint64
		minerFees, err = tx.ComputeFees()
		if err != nil {
			return
		}

		var selectedFeeToken *string
		var selectedFee uint64

		for token := range fees.FEES_PER_BYTE {
			if minerFees[token] != 0 {
				feePerByte := minerFees[token] / tx.Bloom.Size
				if feePerByte >= fees.FEES_PER_BYTE[token] {
					selectedFeeToken = &token
					selectedFee = minerFees[*selectedFeeToken]
					break
				}
			}
		}

		//if it is mine and no fee was paid, let's fake a fee
		if mine && selectedFeeToken == nil {
			selectedFeeToken = &config.NATIVE_TOKEN_STRING
			selectedFee = fees.FEES_PER_BYTE[config.NATIVE_TOKEN_STRING]
		}

		if selectedFeeToken == nil {
			gui.Error("Transaction fee was not accepted")
		} else {
			finalTxs = append(finalTxs, &mempoolTx{
				Tx:          tx,
				Added:       time.Now().Unix(),
				FeePerByte:  selectedFee / tx.Bloom.Size,
				FeeToken:    []byte(*selectedFeeToken),
				Mine:        mine,
				ChainHeight: height,
			})
		}

	}

	if len(finalTxs) == 0 {
		return false, errors.New("Transactions don't meet the criteria")
	}

	mempool.txs.txsListMutex.Lock()
	defer mempool.txs.txsListMutex.Unlock()

	list := mempool.txs.txsList.Load().([]*mempoolTx)

	var txsCount int64
	for _, newTx := range finalTxs {

		found := false
		for _, tx2 := range list {
			if bytes.Equal(tx2.Tx.Bloom.Hash, newTx.Tx.Bloom.Hash) {
				found = true
				break
			}
		}

		if !found {

			//making sure that the transaction is not inserted twice
			txsCount = atomic.AddInt64(&mempool.txs.txsCount, 1)
			atomic.AddInt64(&mempool.txs.txsInserted, 1)

			//appending
			list = append(list, newTx)
			mempool.txs.txsList.Store(list)

		}

	}

	gui.Info2Update("mempool", strconv.FormatInt(txsCount, 10))

	return true, nil
}

func (mempool *Mempool) Exists(txId []byte) bool {
	list := mempool.txs.txsList.Load().([]*mempoolTx)
	for _, tx := range list {
		if bytes.Equal(tx.Tx.Bloom.Hash, txId) {
			return true
		}
	}
	return false
}

func (mempool *Mempool) Delete(txId []byte) *transaction.Transaction {

	if !mempool.Exists(txId) {
		return nil
	}

	mempool.txs.txsListMutex.Lock()
	defer mempool.txs.txsListMutex.Unlock()

	list := mempool.txs.txsList.Load().([]*mempoolTx)
	for i, tx := range list {
		if bytes.Equal(tx.Tx.Bloom.Hash, txId) {
			list = append(list[:i], list[i+1:]...)

			txsCount := atomic.AddInt64(&mempool.txs.txsCount, -1)
			gui.Info2Update("mempool", strconv.FormatInt(txsCount, 10))

			return tx.Tx
		}
	}

	return nil
}

//reset the forger
func (mempool *Mempool) UpdateWork(hash []byte, height uint64) {
	mempool.newWork <- &mempoolWork{
		chainHash:   hash,
		chainHeight: height,
	}
}
func (mempool *Mempool) RestartWork() {
	mempool.newWork <- nil
}

func InitMemPool() (mempool *Mempool, err error) {

	gui.Log("MemPool init...")

	txsList := atomic.Value{}
	txsList.Store([]*mempoolTx{})

	mempool = &Mempool{
		newWork: make(chan *mempoolWork),
		result:  &mempoolResult{},
		txs: &mempoolTxs{
			txsList: txsList,
		},
	}

	go func() {
		for {
			mempool.print()
			time.Sleep(60 * time.Second)
		}
	}()

	worker := new(mempoolWorker)
	go worker.processing(mempool.newWork, mempool.txs, mempool.result)

	return
}
