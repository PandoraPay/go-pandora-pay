package mempool

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/config/fees"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

type mempoolTx struct {
	Tx          *transaction.Transaction
	Added       int64
	Mine        bool
	FeePerByte  uint64
	FeeToken    []byte //20 byte
	ChainHeight uint64
}

type mempoolTxs struct {
	txsCount     uint64
	txsMap       sync.Map
	txsList      []*mempoolTx //we use the list because there are less operations of inserting into the mempool. Txs are coming less frequent
	txsInserted  uint16
	sync.RWMutex `json:"-"`
}

type Mempool struct {
	txs        mempoolTxs
	updateTask unsafe.Pointer // *memPoolUpdateTask used to know when the work has been changed
	result     *mempoolResult
}

func (mempool *Mempool) AddTxToMemPoolSilent(tx *transaction.Transaction, height uint64, mine bool) (result bool, err error) {
	defer func() {
		err = helpers.ConvertRecoverError(recover())
	}()
	result = mempool.AddTxToMemPool(tx, height, mine)
	return
}

func (mempool *Mempool) AddTxToMemPool(tx *transaction.Transaction, height uint64, mine bool) bool {

	tx.VerifyBloomAll()
	if _, found := mempool.txs.txsMap.Load(tx.Bloom.HashStr); found {
		return false
	}

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
		selectedFeeToken = &config.NATIVE_TOKEN_STRING
		selectedFee = fees.FEES_PER_BYTE[config.NATIVE_TOKEN_STRING]
	}

	if selectedFeeToken == nil {
		panic("Transaction fee was not accepted")
	}

	mempoolTx := &mempoolTx{
		Tx:          tx,
		Added:       time.Now().Unix(),
		FeePerByte:  selectedFee / size,
		FeeToken:    []byte(*selectedFeeToken),
		Mine:        mine,
		ChainHeight: height,
	}

	//meanwhile it was inserted, if not, let's store it
	if _, exists := mempool.txs.txsMap.LoadOrStore(tx.Bloom.HashStr, mempoolTx); exists {
		return false
	}

	//making sure that the transaction is not inserted twice
	mempool.txs.Lock()
	defer mempool.txs.Unlock()

	mempool.txs.txsCount += 1
	mempool.txs.txsInserted += 1
	mempool.txs.txsMap.Store(tx.Bloom.HashStr, mempoolTx)
	mempool.txs.txsList = append(mempool.txs.txsList, mempoolTx)

	gui.Info2Update("mempool", strconv.FormatUint(mempool.txs.txsCount, 10))

	return true
}

func (mempool *Mempool) Exists(txId []byte) bool {
	_, found := mempool.txs.txsMap.Load(string(txId))
	return found
}

func (mempool *Mempool) Delete(txId []byte) (tx *transaction.Transaction) {

	hashStr := string(txId)

	if _, found := mempool.txs.txsMap.Load(hashStr); found == false {
		return nil
	}

	mempool.txs.txsMap.Delete(hashStr)

	mempool.txs.Lock()
	defer mempool.txs.Unlock()

	for i, txOut := range mempool.txs.txsList {
		if txOut.Tx.Bloom.HashStr == hashStr {
			//order is not important
			mempool.txs.txsList[i] = mempool.txs.txsList[len(mempool.txs.txsList)-1]
			mempool.txs.txsList = mempool.txs.txsList[:len(mempool.txs.txsList)-1]
			mempool.txs.txsCount -= 1
			break
		}
	}

	gui.Info2Update("mempool", strconv.FormatUint(mempool.txs.txsCount, 10))
	return
}

func InitMemPool() (mempool *Mempool) {

	gui.Log("MemPool init...")

	mempool = &Mempool{
		updateTask: nil,
		result:     &mempoolResult{},
		txs: mempoolTxs{
			txsList: []*mempoolTx{},
		},
	}

	go func() {
		for {
			mempool.print()
			time.Sleep(60 * time.Second)
		}
	}()

	initWorker(mempool)

	return
}
