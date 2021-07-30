package mempool

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_simple "pandora-pay/blockchain/transactions/transaction/transaction-simple"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config"
	"pandora-pay/config/config_fees"
	"pandora-pay/gui"
	"pandora-pay/helpers/multicast"
	"pandora-pay/recovery"
	"sync/atomic"
	"time"
)

type MempoolTxBroadcastNotification struct {
	Txs              []*transaction.Transaction
	AwaitPropagation bool
	ExceptSocketUUID string
}

type mempoolTx struct {
	Tx          *transaction.Transaction `json:"tx"`
	Added       int64                    `json:"added"`
	Mine        bool                     `json:"mine"`
	FeePerByte  uint64                   `json:"feePerByte"`
	FeeToken    []byte                   `json:"feeToken"` //20 byte
	ChainHeight uint64                   `json:"chainHeight"`
}

type mempoolTxProcess struct {
	tx  *mempoolTx
	err error
}

type Mempool struct {
	result                  *atomic.Value                `json:"-"` //*MempoolResult
	SuspendProcessingCn     chan struct{}                `json:"-"`
	ContinueProcessingCn    chan bool                    `json:"-"`
	newWorkCn               chan *mempoolWork            `json:"-"`
	addTransactionCn        chan *MempoolWorkerAddTx     `json:"-"`
	removeTransactionsCn    chan *MempoolWorkerRemoveTxs `json:"-"`
	insertTransactionsCn    chan *MempoolWorkerInsertTxs `json:"-"`
	Txs                     *MempoolTxs                  `json:"-"`
	Wallet                  *mempoolWallet               `json:"-"`
	NewTransactionMulticast *multicast.MulticastChannel  `json:"-"`
}

func (mempool *Mempool) RemoveInsertedTxsFromBlockchain(txs []*transaction.Transaction) bool {
	answerCn := make(chan bool)
	mempool.removeTransactionsCn <- &MempoolWorkerRemoveTxs{txs, answerCn}
	return <-answerCn
}

func (mempool *Mempool) InsertRemovedTxsFromBlockchain(txs []*transaction.Transaction, height uint64) bool {
	finalTxs := mempool.processTxsToMemPool(txs, height)

	insertTxs := make([]*mempoolTx, len(finalTxs))
	for i, it := range finalTxs {
		if it != nil {
			insertTxs[i] = it.tx
		}
	}

	answerCn := make(chan bool)
	mempool.insertTransactionsCn <- &MempoolWorkerInsertTxs{insertTxs, answerCn}
	return <-answerCn

}

func (mempool *Mempool) AddTxToMemPool(tx *transaction.Transaction, height uint64, awaitAnswer, awaitBroadcasting bool, exceptSocketUUID string) error {
	result := mempool.AddTxsToMemPool([]*transaction.Transaction{tx}, height, awaitAnswer, awaitBroadcasting, exceptSocketUUID)
	return result[0]
}

func (mempool *Mempool) processTxsToMemPool(txs []*transaction.Transaction, height uint64) []*mempoolTxProcess {

	finalTxs := make([]*mempoolTxProcess, len(txs))

	for i, tx := range txs {

		finalTxs[i] = &mempoolTxProcess{}

		if err := tx.VerifyBloomAll(); err != nil {
			finalTxs[i].err = err
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
				if mempool.Wallet.Exists(vin.Bloom.PublicKeyHash) {
					mine = true
					break
				}
			}
		}

		minerFees, err := tx.GetAllFees()
		if err != nil {
			finalTxs[i].err = err
			continue
		}

		var selectedFeeToken *string
		var selectedFee uint64

		for token := range config_fees.FEES_PER_BYTE {
			if minerFees[token] != 0 {
				feePerByte := minerFees[token] / tx.Bloom.Size
				if feePerByte >= config_fees.FEES_PER_BYTE[token] {
					selectedFeeToken = &token
					selectedFee = minerFees[*selectedFeeToken]
					break
				}
			}
		}

		//if it is mine and no fee was paid, let's fake a fee
		if mine && selectedFeeToken == nil {
			selectedFeeToken = &config.NATIVE_TOKEN_STRING
			selectedFee = config_fees.FEES_PER_BYTE[config.NATIVE_TOKEN_STRING]
		}

		if selectedFeeToken == nil {
			finalTxs[i].err = errors.New("Transaction fee was not accepted")
			continue
		} else {
			finalTxs[i].tx = &mempoolTx{
				Tx:          tx,
				Added:       time.Now().Unix(),
				FeePerByte:  selectedFee / tx.Bloom.Size,
				FeeToken:    []byte(*selectedFeeToken),
				Mine:        mine,
				ChainHeight: height,
			}
		}

	}

	return finalTxs
}

func (mempool *Mempool) AddTxsToMemPool(txs []*transaction.Transaction, height uint64, awaitAnswer, awaitBroadcasting bool, exceptSocketUUID string) []error {

	finalTxs := mempool.processTxsToMemPool(txs, height)

	//making sure that the transaction is not inserted twice
	for _, finalTx := range finalTxs {
		if finalTx.tx != nil {

			var errorResult error

			_, loaded := mempool.Txs.txs.LoadOrStore(finalTx.tx.Tx.Bloom.HashStr, finalTx.tx.Tx)
			if loaded {
				errorResult = errors.New("Tx already exists")
			} else if awaitAnswer {
				answerCn := make(chan error)
				mempool.addTransactionCn <- &MempoolWorkerAddTx{finalTx.tx, answerCn}
				errorResult = <-answerCn
			} else {
				mempool.addTransactionCn <- &MempoolWorkerAddTx{finalTx.tx, nil}
			}

			if errorResult != nil {
				finalTx.err = errorResult
				finalTx.tx = nil
			}

		}
	}

	if exceptSocketUUID != "*" {

		notNull := 0
		for _, finalTx := range finalTxs {
			if finalTx.tx != nil {
				notNull += 1
			}
		}
		broadcastTxs := make([]*transaction.Transaction, notNull)

		notNull = 0
		for _, finalTx := range finalTxs {
			if finalTx.tx != nil {
				broadcastTxs[notNull] = finalTx.tx.Tx
				notNull += 1
			}
		}

		mempool.NewTransactionMulticast.BroadcastAwait(&MempoolTxBroadcastNotification{
			broadcastTxs,
			awaitBroadcasting,
			exceptSocketUUID,
		})

	}

	out := make([]error, len(txs))
	for i, finalTx := range finalTxs {
		out[i] = finalTx.err
	}
	return out
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

	newWork := &mempoolWork{
		chainHash:    hash,
		chainHeight:  height,
		result:       result,
		waitAnswerCn: make(chan struct{}),
	}

	mempool.newWorkCn <- newWork
	<-newWork.waitAnswerCn

}

func (mempool *Mempool) ContinueWork() {
	newWork := &mempoolWork{
		waitAnswerCn: make(chan struct{}),
	}
	mempool.newWorkCn <- newWork
	<-newWork.waitAnswerCn
}

func CreateMemPool() (*Mempool, error) {

	gui.GUI.Log("MemPool init...")

	mempool := &Mempool{
		result:                  &atomic.Value{}, // *MempoolResult
		Txs:                     createMempoolTxs(),
		SuspendProcessingCn:     make(chan struct{}),
		ContinueProcessingCn:    make(chan bool),
		newWorkCn:               make(chan *mempoolWork),
		addTransactionCn:        make(chan *MempoolWorkerAddTx),
		removeTransactionsCn:    make(chan *MempoolWorkerRemoveTxs),
		insertTransactionsCn:    make(chan *MempoolWorkerInsertTxs),
		Wallet:                  createMempoolWallet(),
		NewTransactionMulticast: multicast.NewMulticastChannel(),
	}

	worker := new(mempoolWorker)
	recovery.SafeGo(func() {
		worker.processing(mempool.newWorkCn, mempool.SuspendProcessingCn, mempool.ContinueProcessingCn, mempool.addTransactionCn, mempool.insertTransactionsCn, mempool.removeTransactionsCn, mempool.Txs)
	})

	mempool.initCLI()

	return mempool, nil
}
