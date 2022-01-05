package mempool

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/config/config_fees"
	"pandora-pay/gui"
	"pandora-pay/helpers"
	"pandora-pay/helpers/generics"
	"pandora-pay/network/websocks/connection/advanced_connection_types"
	"pandora-pay/recovery"
	"runtime"
	"time"
)

type mempoolTx struct {
	Tx          *transaction.Transaction `json:"tx"`
	Added       int64                    `json:"added"`
	Mine        bool                     `json:"mine"`
	FeePerByte  uint64                   `json:"feePerByte"`
	ChainHeight uint64                   `json:"chainHeight"`
}

type Mempool struct {
	result                    *generics.Value[*MempoolResult] `json:"-"`
	SuspendProcessingCn       chan struct{}                   `json:"-"`
	ContinueProcessingCn      chan ContinueProcessingType     `json:"-"`
	newWorkCn                 chan *mempoolWork               `json:"-"`
	addTransactionCn          chan *MempoolWorkerAddTx        `json:"-"`
	removeTransactionsCn      chan *MempoolWorkerRemoveTxs    `json:"-"`
	insertTransactionsCn      chan *MempoolWorkerInsertTxs    `json:"-"`
	Txs                       *MempoolTxs                     `json:"-"`
	OnBroadcastNewTransaction func([]*transaction.Transaction, bool, bool, advanced_connection_types.UUID) []error
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

	finalTxs, _ := mempool.processTxsToMempool(txs, height)

	insertTxs := make([]*mempoolTx, len(finalTxs))
	for i, it := range finalTxs {
		if it != nil {
			insertTxs[i] = it
		}
	}

	answerCn := make(chan bool)
	mempool.insertTransactionsCn <- &MempoolWorkerInsertTxs{insertTxs, answerCn}
	return <-answerCn

}

func (mempool *Mempool) AddTxToMempool(tx *transaction.Transaction, height uint64, justCreated bool, awaitAnswer, awaitBroadcasting bool, exceptSocketUUID advanced_connection_types.UUID) error {
	result := mempool.AddTxsToMempool([]*transaction.Transaction{tx}, height, justCreated, awaitAnswer, awaitBroadcasting, exceptSocketUUID)
	return result[0]
}

func (mempool *Mempool) processTxsToMempool(txs []*transaction.Transaction, height uint64) (finalTxs []*mempoolTx, errs []error) {

	finalTxs = make([]*mempoolTx, len(txs))
	errs = make([]error, len(txs))

	for i, tx := range txs {

		if err := tx.VerifyBloomAll(); err != nil {
			errs[i] = err
			continue
		}

		if mempool.Txs.Exists(tx.Bloom.HashStr) {
			continue
		}

		minerFee, err := tx.GetAllFee()
		if err != nil {
			errs[i] = err
			continue
		}

		computedFeePerByte := minerFee
		if err = helpers.SafeUint64Sub(&computedFeePerByte, tx.SpaceExtra*config_fees.FEE_PER_BYTE_EXTRA_SPACE); err != nil {
			errs[i] = err
			continue
		}

		computedFeePerByte = minerFee / tx.Bloom.Size

		requiredFeePerByte := uint64(0)
		switch tx.Version {
		case transaction_type.TX_SIMPLE:
			requiredFeePerByte = config_fees.FEE_PER_BYTE
		case transaction_type.TX_ZETHER:
			requiredFeePerByte = config_fees.FEE_PER_BYTE_ZETHER
		default:
			errs[i] = errors.New("Invalid Tx.Version")
			continue
		}

		if computedFeePerByte < requiredFeePerByte {
			errs[i] = errors.New("Transaction fee was not accepted")
			continue
		}

		finalTxs[i] = &mempoolTx{
			Tx:          tx,
			Added:       time.Now().Unix(),
			FeePerByte:  computedFeePerByte,
			ChainHeight: height,
		}

	}

	return
}

func (mempool *Mempool) AddTxsToMempool(txs []*transaction.Transaction, height uint64, justCreated, awaitAnswer, awaitBroadcasting bool, exceptSocketUUID advanced_connection_types.UUID) []error {

	finalTxs, errs := mempool.processTxsToMempool(txs, height)

	//making sure that the transaction is not inserted twice
	if runtime.GOARCH != "wasm" {
		for i, finalTx := range finalTxs {
			if finalTx != nil {

				var errorResult error

				inserted := mempool.Txs.insertTx(finalTx)
				if !inserted {
					errorResult = errors.New("Tx already exists")
				} else if awaitAnswer {
					answerCn := make(chan error)
					mempool.addTransactionCn <- &MempoolWorkerAddTx{finalTx, answerCn}
					errorResult = <-answerCn
				} else {
					mempool.addTransactionCn <- &MempoolWorkerAddTx{finalTx, nil}
				}

				if errorResult != nil {
					errs[i] = errorResult
					finalTxs[i] = nil
				}

			}
		}
	}

	if exceptSocketUUID != advanced_connection_types.UUID_SKIP_ALL {

		broadcastTxs := make([]*transaction.Transaction, len(finalTxs))
		for i, finalTx := range finalTxs {
			if finalTx != nil {
				broadcastTxs[i] = finalTx.Tx
			}
		}

		errors2 := mempool.OnBroadcastNewTransaction(broadcastTxs, justCreated, awaitBroadcasting, exceptSocketUUID)
		for i, err := range errors2 {
			if err != nil {
				errs[i] = err
				finalTxs[i] = nil
			}
		}

	}

	return errs
}

//reset the forger
func (mempool *Mempool) UpdateWork(hash []byte, height uint64) {

	result := &MempoolResult{
		txs:         &generics.Value[[]*mempoolTx]{}, //, appendOnly
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

func CreateMempool() (*Mempool, error) {

	gui.GUI.Log("Mempool init...")

	mempool := &Mempool{
		result:               &generics.Value[*MempoolResult]{}, // *MempoolResult
		Txs:                  createMempoolTxs(),
		SuspendProcessingCn:  make(chan struct{}),
		ContinueProcessingCn: make(chan ContinueProcessingType),
		newWorkCn:            make(chan *mempoolWork),
		addTransactionCn:     make(chan *MempoolWorkerAddTx, 1000),
		removeTransactionsCn: make(chan *MempoolWorkerRemoveTxs),
		insertTransactionsCn: make(chan *MempoolWorkerInsertTxs),
	}

	worker := new(mempoolWorker)
	recovery.SafeGo(func() {
		worker.processing(mempool.newWorkCn, mempool.SuspendProcessingCn, mempool.ContinueProcessingCn, mempool.addTransactionCn, mempool.insertTransactionsCn, mempool.removeTransactionsCn, mempool.Txs)
	})

	mempool.initCLI()

	return mempool, nil
}
