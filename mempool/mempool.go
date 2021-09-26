package mempool

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	transaction_type "pandora-pay/blockchain/transactions/transaction/transaction-type"
	"pandora-pay/config/config_fees"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/multicast"
	"pandora-pay/network/websocks/connection/advanced-connection-types"
	"pandora-pay/recovery"
	"sync/atomic"
	"time"
)

type MempoolTxBroadcastNotification struct {
	Txs              []*transaction.Transaction
	AwaitPropagation bool
	ExceptSocketUUID advanced_connection_types.UUID
}

type mempoolTx struct {
	Tx          *transaction.Transaction `json:"tx"`
	Added       int64                    `json:"added"`
	Mine        bool                     `json:"mine"`
	FeePerByte  uint64                   `json:"feePerByte"`
	ChainHeight uint64                   `json:"chainHeight"`
}

type mempoolTxProcess struct {
	tx  *mempoolTx
	err error
}

type Mempool struct {
	result                  *atomic.Value                `json:"-"` //*MempoolResult
	SuspendProcessingCn     chan struct{}                `json:"-"`
	ContinueProcessingCn    chan ContinueProcessingType  `json:"-"`
	newWorkCn               chan *mempoolWork            `json:"-"`
	addTransactionCn        chan *MempoolWorkerAddTx     `json:"-"`
	removeTransactionsCn    chan *MempoolWorkerRemoveTxs `json:"-"`
	insertTransactionsCn    chan *MempoolWorkerInsertTxs `json:"-"`
	Txs                     *MempoolTxs                  `json:"-"`
	NewTransactionMulticast *multicast.MulticastChannel  `json:"-"`
}

func (mempool *Mempool) ContinueProcessing(continueProcessingType ContinueProcessingType) {
	mempool.ContinueProcessingCn <- continueProcessingType
}

func (mempool *Mempool) RemoveInsertedTxsFromBlockchain(txs []string) bool {
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

func (mempool *Mempool) AddTxToMemPool(tx *transaction.Transaction, height uint64, awaitAnswer, awaitBroadcasting bool, exceptSocketUUID advanced_connection_types.UUID) error {
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

		if mempool.Txs.Exists(tx.Bloom.HashStr) {
			continue
		}

		minerFees, err := tx.GetAllFees()
		if err != nil {
			finalTxs[i].err = err
			continue
		}

		computedFeePerByte := minerFees
		if err = helpers.SafeUint64Sub(&computedFeePerByte, uint64(64*len(tx.Registrations.Registrations))*config_fees.FEES_PER_BYTE_EXTRA_SPACE); err != nil {
			finalTxs[i].err = err
			continue
		}

		computedFeePerByte = minerFees / tx.Bloom.Size

		requiredFeePerByte := uint64(0)
		switch tx.Version {
		case transaction_type.TX_SIMPLE:
			requiredFeePerByte = config_fees.FEES_PER_BYTE
		case transaction_type.TX_ZETHER:
			requiredFeePerByte = config_fees.FEES_PER_BYTE_ZETHER
		default:
			finalTxs[i].err = errors.New("Invalid Tx.Version")
			continue
		}

		if computedFeePerByte < requiredFeePerByte {
			finalTxs[i].err = errors.New("Transaction fee was not accepted")
			continue
		}

		finalTxs[i].tx = &mempoolTx{
			Tx:          tx,
			Added:       time.Now().Unix(),
			FeePerByte:  computedFeePerByte,
			ChainHeight: height,
		}

	}

	return finalTxs
}

func (mempool *Mempool) AddTxsToMemPool(txs []*transaction.Transaction, height uint64, awaitAnswer, awaitBroadcasting bool, exceptSocketUUID advanced_connection_types.UUID) []error {

	finalTxs := mempool.processTxsToMemPool(txs, height)

	//making sure that the transaction is not inserted twice
	for _, finalTx := range finalTxs {
		if finalTx.tx != nil {

			var errorResult error

			inserted := mempool.Txs.insertTx(finalTx.tx)
			if !inserted {
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

	if exceptSocketUUID != advanced_connection_types.UUID_SKIP_ALL {

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
		chainHash:   hash,
		chainHeight: height,
		result:      result,
	}

	mempool.newWorkCn <- newWork
}

func (mempool *Mempool) ContinueWork() {
	newWork := &mempoolWork{}
	mempool.newWorkCn <- newWork
}

func CreateMemPool() (*Mempool, error) {

	gui.GUI.Log("MemPool init...")

	mempool := &Mempool{
		result:                  &atomic.Value{}, // *MempoolResult
		Txs:                     createMempoolTxs(),
		SuspendProcessingCn:     make(chan struct{}),
		ContinueProcessingCn:    make(chan ContinueProcessingType),
		newWorkCn:               make(chan *mempoolWork),
		addTransactionCn:        make(chan *MempoolWorkerAddTx, 1000),
		removeTransactionsCn:    make(chan *MempoolWorkerRemoveTxs),
		insertTransactionsCn:    make(chan *MempoolWorkerInsertTxs),
		NewTransactionMulticast: multicast.NewMulticastChannel(),
	}

	worker := new(mempoolWorker)
	recovery.SafeGo(func() {
		worker.processing(mempool.newWorkCn, mempool.SuspendProcessingCn, mempool.ContinueProcessingCn, mempool.addTransactionCn, mempool.insertTransactionsCn, mempool.removeTransactionsCn, mempool.Txs)
	})

	mempool.initCLI()

	return mempool, nil
}
