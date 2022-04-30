package txs_validator

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/gui"
	"pandora-pay/helpers/generics"
	"sync/atomic"
	"time"
)

type TxsValidator struct {
	all                 *generics.Map[string, *txValidatedWork]
	workers             []*TxsValidatorWorker
	newValidationWorkCn chan *txValidatedWork
}

func (validator *TxsValidator) MarkAsValidatedTx(tx *transaction.Transaction) error {

	foundWork, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidatedWork{make(chan struct{}), TX_VALIDATED_INIT, tx, 0, nil, nil})

	if !loaded {
		if err := foundWork.tx.BloomAll(); err != nil {
			foundWork.result = err
		}
		foundWork.bloomExtra = foundWork.tx.TransactionBaseInterface.GetBloomExtra()
		foundWork.tx = nil
		foundWork.time = time.Now().Add(EXPIRE_TIME_MS).Unix()
		atomic.StoreInt32(&foundWork.status, TX_VALIDATED_PROCCESSED)
		close(foundWork.wait)
		return nil
	} else {

		if atomic.LoadInt32(&foundWork.status) == TX_VALIDATED_PROCCESSED {
			if foundWork.result != nil {
				gui.GUI.Error("Strange Error. FoundWork.result is false")
				return foundWork.result
			}
			tx.TransactionBaseInterface.SetBloomExtra(foundWork.bloomExtra)
			return nil
		}

		return tx.BloomAll()
	}

}

//blocking
func (validator *TxsValidator) ValidateTx(tx *transaction.Transaction) error {

	foundWork, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidatedWork{make(chan struct{}), TX_VALIDATED_INIT, tx, 0, nil, nil})
	if !loaded {
		validator.newValidationWorkCn <- foundWork
	}

	<-foundWork.wait
	if foundWork.result != nil {
		return foundWork.result
	}

	tx.TransactionBaseInterface.SetBloomExtra(foundWork.bloomExtra)

	return nil
}

func (validator *TxsValidator) ValidateTxs(txs []*transaction.Transaction) error {

	outputs := make([]*txValidatedWork, len(txs))
	for i, tx := range txs {
		foundWork, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidatedWork{make(chan struct{}), TX_VALIDATED_INIT, tx, 0, nil, nil})
		if !loaded {
			validator.newValidationWorkCn <- foundWork
		}
		outputs[i] = foundWork
	}

	for _, foundWork := range outputs {
		<-foundWork.wait
		if foundWork.result != nil {
			return foundWork.result
		}
	}

	for i, foundWork := range outputs {
		txs[i].TransactionBaseInterface.SetBloomExtra(foundWork.bloomExtra)
	}

	return nil
}

func (validator *TxsValidator) runRemoveExpiredTransactions() {

	c := 0
	for {

		now := time.Now().Unix()

		validator.all.Range(func(key string, work *txValidatedWork) bool {

			if atomic.LoadInt32(&work.status) == TX_VALIDATED_PROCCESSED {

				if work.time < now {
					validator.all.Delete(key)
				}

				c += 1
				if c%200 == 0 {
					time.Sleep(50 * time.Millisecond)
				}
			}

			return true
		})

		time.Sleep(time.Second)
	}
}

func NewTxsValidator() (*TxsValidator, error) {

	threadsCount := 10 * config.CPU_THREADS
	if config.LIGHT_COMPUTATIONS {
		threadsCount = generics.Max(1, config.CPU_THREADS/2)
	}

	txsValidator := &TxsValidator{
		&generics.Map[string, *txValidatedWork]{},
		make([]*TxsValidatorWorker, threadsCount),
		make(chan *txValidatedWork, 1),
	}

	for i := range txsValidator.workers {
		txsValidator.workers[i] = newTxsValidatorWorker(txsValidator.newValidationWorkCn)
	}

	for _, worker := range txsValidator.workers {
		worker.start()
	}

	go txsValidator.runRemoveExpiredTransactions()

	return txsValidator, nil
}
