package txs_validator

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
	"sync/atomic"
	"time"
)

type TxsValidator struct {
	all                 *generics.Map[string, *txValidated]
	workers             []*TxsValidatorWorker
	newValidationWorkCn chan *txValidated
}

func (validator *TxsValidator) MarkAsValidatedTx(tx *transaction.Transaction) error {

	foundWork, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidated{make(chan struct{}), TX_VALIDATED_INIT, tx, 0, nil, nil})
	if !loaded {
		if err := foundWork.tx.BloomAll(); err != nil {
			foundWork.result = err
		}
		foundWork.bloomExtra = foundWork.tx.TransactionBaseInterface.GetBloomExtra()
		foundWork.tx = nil
		atomic.StoreInt32(&foundWork.status, TX_VALIDATED_PROCCESSED)
		foundWork.time = time.Now().Add(EXPIRE_TIME_MS).Unix()
		close(foundWork.wait)
	}

	<-foundWork.wait
	if err := foundWork.result; err != nil {
		return err
	}

	tx.TransactionBaseInterface.SetBloomExtra(foundWork.bloomExtra)

	return nil
}

//blocking
func (validator *TxsValidator) ValidateTx(tx *transaction.Transaction) error {

	foundWork, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidated{make(chan struct{}), TX_VALIDATED_INIT, tx, 0, nil, nil})
	if !loaded {
		validator.newValidationWorkCn <- foundWork
	}

	<-foundWork.wait
	if err := foundWork.result; err != nil {
		return err
	}

	tx.TransactionBaseInterface.SetBloomExtra(foundWork.bloomExtra)

	return nil
}

func (validator *TxsValidator) ValidateTxs(txs []*transaction.Transaction) error {

	outputs := make([]*txValidated, len(txs))
	for i, tx := range txs {
		foundWork, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidated{make(chan struct{}), TX_VALIDATED_INIT, tx, 0, nil, nil})
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

	ticker := time.NewTicker(100 * time.Millisecond)

	c := 0
	for {

		<-ticker.C

		now := time.Now().Unix()

		validator.all.Range(func(key string, work *txValidated) bool {

			if atomic.LoadInt32(&work.status) == TX_VALIDATED_PROCCESSED {

				if work.time < now {
					validator.all.Delete(key)
				}

				c += 1
				if c%1000 == 0 {
					<-ticker.C
				}
			}

			return true
		})

	}
}

func NewTxsValidator() (*TxsValidator, error) {

	txsValidator := &TxsValidator{
		&generics.Map[string, *txValidated]{},
		nil,
		make(chan *txValidated),
	}

	workers := make([]*TxsValidatorWorker, config.CPU_THREADS)
	for i := range workers {
		workers[i] = newTxsValidatorWorker(txsValidator.newValidationWorkCn)
	}

	txsValidator.workers = workers

	for _, worker := range workers {
		worker.start()
	}

	go txsValidator.runRemoveExpiredTransactions()

	return txsValidator, nil
}
