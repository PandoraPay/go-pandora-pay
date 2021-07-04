package mempool

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/config/fees"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/recovery"
	"sync/atomic"
	"time"
)

type mempoolTx struct {
	Tx          *transaction.Transaction `json:"tx"`
	Added       int64                    `json:"added"`
	Mine        bool                     `json:"mine"`
	FeePerByte  uint64                   `json:"feePerByte"`
	FeeToken    []byte                   `json:"feeToken"` //20 byte
	ChainHeight uint64                   `json:"chainHeight"`
}

type Mempool struct {
	result                  *atomic.Value               `json:"-"` //*MempoolResult
	SuspendProcessingCn     chan struct{}               `json:"-"`
	ContinueProcessingCn    chan struct{}               `json:"-"`
	NewWorkCn               chan *mempoolWork           `json:"-"`
	AddTransactionCn        chan *MempoolWorkerAddTx    `json:"-"`
	Txs                     *MempoolTxs                 `json:"-"`
	Wallet                  *mempoolWallet              `json:"-"`
	NewTransactionMulticast *multicast.MulticastChannel `json:"-"`
}

func (mempool *Mempool) AddTxToMemPool(tx *transaction.Transaction, height uint64, propagateToSockets, awaitAnswer bool) (bool, error) {
	result, err := mempool.AddTxsToMemPool([]*transaction.Transaction{tx}, height, propagateToSockets, awaitAnswer)
	if err != nil {
		return false, err
	}
	return result[0], nil
}

func (mempool *Mempool) processTxsToMemPool(txs []*transaction.Transaction, height uint64) (bool, []*mempoolTx, error) {

	finalTxs := make([]*mempoolTx, len(txs))

	mempool.Wallet.Lock()
	defer mempool.Wallet.Unlock()

	for i, tx := range txs {

		if err := tx.VerifyBloomAll(); err != nil {
			continue
		}

		if mempool.Txs.Exists(tx.Bloom.HashStr) != nil {
			continue
		}

		mine := false

		switch tx.TxType {
		case transaction_type.TX_SIMPLE:
			txBase := tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
			for _, vin := range txBase.Vin {
				if mempool.Wallet.myAddressesMap[string(vin.Bloom.PublicKeyHash)] != nil {
					mine = true
					break
				}
			}
		}

		minerFees, err := tx.GetAllFees()
		if err != nil {
			return false, nil, err
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
			return false, nil, errors.New("Transaction fee was not accepted")
		} else {
			finalTxs[i] = &mempoolTx{
				Tx:          tx,
				Added:       time.Now().Unix(),
				FeePerByte:  selectedFee / tx.Bloom.Size,
				FeeToken:    []byte(*selectedFeeToken),
				Mine:        mine,
				ChainHeight: height,
			}
		}

	}

	return true, finalTxs, nil
}

func (mempool *Mempool) AddTxsToMemPool(txs []*transaction.Transaction, height uint64, propagateToSockets, awaitAnswer bool) ([]bool, error) {

	_, finalTxs, err := mempool.processTxsToMemPool(txs, height)
	if err != nil {
		return nil, err
	}

	//making sure that the transaction is not inserted twice
	out := make([]bool, len(finalTxs))
	for i, tx := range finalTxs {
		if tx != nil {

			if awaitAnswer {
				answerCn := make(chan bool)
				mempool.AddTransactionCn <- &MempoolWorkerAddTx{tx, answerCn}
				out[i] = <-answerCn
			} else {
				mempool.AddTransactionCn <- &MempoolWorkerAddTx{tx, nil}
				out[i] = true
			}
		} else {
			out[i] = false
		}
	}

	if propagateToSockets {
		for i, result := range out {
			if result {
				mempool.NewTransactionMulticast.BroadcastAwait(finalTxs[i].Tx)
			}
		}
	}

	return out, nil
}

//reset the forger
func (mempool *Mempool) UpdateWork(hash []byte, height uint64) {

	result := &MempoolResult{
		txs:         &atomic.Value{}, //[]*mempoolTx{} , appendOnly
		totalSize:   0,
		chainHash:   hash,
		chainHeight: height,
	}
	result.txs.Store([]*mempoolTx{})

	mempool.result.Store(result)

	mempool.NewWorkCn <- &mempoolWork{
		chainHash:   hash,
		chainHeight: height,
		result:      result,
	}

}

func CreateMemPool() (*Mempool, error) {

	gui.GUI.Log("MemPool init...")

	mempool := &Mempool{
		result:                  &atomic.Value{}, // *MempoolResult
		Txs:                     createMempoolTxs(),
		SuspendProcessingCn:     make(chan struct{}),
		ContinueProcessingCn:    make(chan struct{}),
		NewWorkCn:               make(chan *mempoolWork),
		AddTransactionCn:        make(chan *MempoolWorkerAddTx),
		Wallet:                  createMempoolWallet(),
		NewTransactionMulticast: multicast.NewMulticastChannel(),
	}

	worker := new(mempoolWorker)
	recovery.SafeGo(func() {
		worker.processing(mempool.NewWorkCn, mempool.SuspendProcessingCn, mempool.ContinueProcessingCn, mempool.AddTransactionCn, mempool.Txs.addToListCn, mempool.Txs.removeFromListCn, mempool.Txs.clearListCn)
	})

	mempool.initCLI()

	return mempool, nil
}
