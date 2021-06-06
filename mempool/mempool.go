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
	ContinueProcessingCn    chan *mempoolWork           `json:"-"`
	AddTransactionCn        chan *MempoolWorkerAddTx    `json:"-"`
	Txs                     *MempoolTxs                 `json:"-"`
	Wallet                  *mempoolWallet              `json:"-"`
	NewTransactionMulticast *multicast.MulticastChannel `json:"-"`
}

func (mempool *Mempool) AddTxToMemPool(tx *transaction.Transaction, height uint64, propagateToSockets bool) (out bool, err error) {
	result, err := mempool.AddTxsToMemPool([]*transaction.Transaction{tx}, height, propagateToSockets)
	return result[0], err
}

func (mempool *Mempool) processTxsToMemPool(txs []*transaction.Transaction, height uint64) (out bool, err error, finalTxs []*mempoolTx) {

	finalTxs = make([]*mempoolTx, len(txs))

	mempool.Wallet.Lock()
	defer mempool.Wallet.Unlock()

	for i, tx := range txs {

		if err = tx.VerifyBloomAll(); err != nil {
			return
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
			return false, errors.New("Transaction fee was not accepted"), nil
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

	out = true
	return
}

func (mempool *Mempool) AddTxsToMemPool(txs []*transaction.Transaction, height uint64, propagateToSockets bool) (out []bool, err error) {

	var finalTxs []*mempoolTx
	if _, err, finalTxs = mempool.processTxsToMemPool(txs, height); err != nil {
		return
	}

	//making sure that the transaction is not inserted twice
	out = make([]bool, len(finalTxs))
	for i, tx := range finalTxs {
		if tx != nil {
			answerCn := make(chan bool)
			mempool.AddTransactionCn <- &MempoolWorkerAddTx{
				Tx:     tx,
				Result: answerCn,
			}
			out[i] = <-answerCn
		} else {
			out[i] = false
		}
	}

	if propagateToSockets {
		for i, result := range out {
			if result {
				mempool.NewTransactionMulticast.Broadcast(finalTxs[i].Tx)
			}
		}
	}

	return
}

//reset the forger
func (mempool *Mempool) UpdateWork(hash []byte, height uint64) {

	result := &MempoolResult{
		txs:         &atomic.Value{},
		totalSize:   0,
		chainHash:   hash,
		chainHeight: height,
	}
	result.txs.Store([]*mempoolTx{})

	mempool.result.Store(result)

	mempool.ContinueProcessingCn <- &mempoolWork{
		chainHash:   hash,
		chainHeight: height,
		result:      result,
	}

}

func CreateMemPool() (mempool *Mempool, err error) {

	gui.GUI.Log("MemPool init...")

	mempool = &Mempool{
		result:                  &atomic.Value{},
		Txs:                     createMempoolTxs(),
		SuspendProcessingCn:     make(chan struct{}),
		ContinueProcessingCn:    make(chan *mempoolWork),
		AddTransactionCn:        make(chan *MempoolWorkerAddTx),
		Wallet:                  createMempoolWallet(),
		NewTransactionMulticast: multicast.NewMulticastChannel(),
	}

	worker := new(mempoolWorker)
	go worker.processing(mempool.SuspendProcessingCn, mempool.ContinueProcessingCn, mempool.AddTransactionCn, mempool.Txs.addToListCn, mempool.Txs.removeFromListCn)

	mempool.initCLI()

	return
}
