package mempool

import (
	"bytes"
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/config/fees"
	"pandora-pay/gui"
	"pandora-pay/helpers"
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
	txs      *atomic.Value //[]*mempoolTx
	txsMutex *sync.Mutex

	txsErrors      *atomic.Value //[]*mempoolTx
	txsErrorsMutex *sync.Mutex

	totalSize uint64 //use atomic

	chainHash   []byte //safe 32
	chainHeight uint64 //safe
}

type mempoolTxs struct {
	txsCount     int64         //use atomic
	txsInserted  int64         //use atomic
	txsList      *atomic.Value // []*mempoolTx
	txsListMutex *sync.Mutex   // for writing
}

type Mempool struct {
	txs                     *mempoolTxs
	result                  *atomic.Value //*mempoolResult
	newWork                 chan *mempoolWork
	Wallet                  *mempoolWallet
	NewTransactionMulticast *helpers.MulticastChannel `json:"-"`
}

func (mempool *Mempool) AddTxToMemPool(tx *transaction.Transaction, height uint64, propagateToSockets bool) (out bool, err error) {
	return mempool.AddTxsToMemPool([]*transaction.Transaction{tx}, height, propagateToSockets)
}

func (mempool *Mempool) AddTxsToMemPool(txs []*transaction.Transaction, height uint64, propagateToSockets bool) (out bool, err error) {

	mempool.Wallet.Lock()
	defer mempool.Wallet.Unlock()

	finalTxs := []*mempoolTx{}

	for _, tx := range txs {

		if err = tx.VerifyBloomAll(); err != nil {
			return
		}

		mine := false

		switch tx.TxType {
		case transaction_type.TxSimple:
			txBase := tx.TxBase.(*transaction_simple.TransactionSimple)
			for _, vin := range txBase.Vin {
				if mempool.Wallet.myAddressesMap[string(vin.Bloom.PublicKeyHash)] != nil {
					mine = true
					break
				}
			}
		}

		var minerFees map[string]uint64
		if minerFees, err = tx.ComputeFees(); err != nil {
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
			return false, errors.New("Transaction fee was not accepted")
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
			atomic.AddInt64(&mempool.txs.txsCount, 1)
			atomic.AddInt64(&mempool.txs.txsInserted, 1)

			//appending
			list = append(list, newTx)
			mempool.txs.txsList.Store(list)

			if propagateToSockets {
				mempool.NewTransactionMulticast.Broadcast(newTx.Tx)
			}

		}

	}

	return true, nil
}

func (mempool *Mempool) Exists(txId []byte) *transaction.Transaction {
	list := mempool.txs.txsList.Load().([]*mempoolTx)
	for _, tx := range list {
		if bytes.Equal(tx.Tx.Bloom.Hash, txId) {
			return tx.Tx
		}
	}
	return nil
}

func (mempool *Mempool) DeleteTx(txId []byte) *transaction.Transaction {
	out := mempool.DeleteTxs([][]byte{txId})
	if len(out) > 0 {
		return out[0]
	}
	return nil
}

func (mempool *Mempool) DeleteTxs(txIds [][]byte) (out []*transaction.Transaction) {

	mempool.txs.txsListMutex.Lock()
	defer mempool.txs.txsListMutex.Unlock()

	list := mempool.txs.txsList.Load().([]*mempoolTx)
	finalList := make([]*mempoolTx, len(list))
	copy(finalList[:], list[:])

	out = []*transaction.Transaction{}

	for _, txId := range txIds {
		for i, tx := range finalList {
			if bytes.Equal(tx.Tx.Bloom.Hash, txId) {

				finalList[i] = finalList[len(finalList)-1]
				finalList = finalList[:len(finalList)-1]

				out = append(out, tx.Tx)
				break
			}
		}
	}

	if len(out) > 0 {
		mempool.txs.txsList.Store(finalList)
		atomic.AddInt64(&mempool.txs.txsCount, -int64(len(out)))
	}

	return
}

//reset the forger
func (mempool *Mempool) UpdateWork(hash []byte, height uint64) {

	result := &mempoolResult{
		txs:            &atomic.Value{},
		txsMutex:       &sync.Mutex{},
		txsErrors:      &atomic.Value{},
		txsErrorsMutex: &sync.Mutex{},
		totalSize:      0,
		chainHash:      hash,
		chainHeight:    height,
	}

	result.txsErrors.Store([]*mempoolTx{})
	result.txs.Store([]*mempoolTx{})

	mempool.result.Store(result)

	mempool.newWork <- &mempoolWork{
		chainHash:   hash,
		chainHeight: height,
		result:      result,
	}
}
func (mempool *Mempool) RestartWork() {
	mempool.newWork <- nil
}

func InitMemPool() (mempool *Mempool, err error) {

	gui.Log("MemPool init...")

	mempool = &Mempool{
		newWork: make(chan *mempoolWork),
		result:  &atomic.Value{},
		txs: &mempoolTxs{
			txsList:      &atomic.Value{},
			txsListMutex: &sync.Mutex{},
		},
		Wallet:                  createMempoolWallet(),
		NewTransactionMulticast: helpers.NewMulticastChannel(),
	}
	mempool.txs.txsList.Store([]*mempoolTx{})

	go func() {
		for {
			mempool.print()
			time.Sleep(60 * time.Second)
		}
	}()

	go func() {
		for {
			gui.Info2Update("mempool", strconv.FormatInt(atomic.LoadInt64(&mempool.txs.txsCount), 10))
			time.Sleep(1 * time.Second)
		}
	}()

	worker := new(mempoolWorker)
	go worker.processing(mempool.newWork, mempool.txs)

	mempool.initCLI()

	return
}
