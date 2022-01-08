package txs_validator

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
	"sync/atomic"
	"time"
)

type TxsValidator struct {
	all        *generics.Map[string, *txValidated]
	processing *generics.Map[string, *txValidated]
	workers    []*TxsValidatorWorker
}

func (validator *TxsValidator) MarkAsValidatedTx(tx *transaction.Transaction) error {

	if err := tx.BloomAll(); err != nil {
		return err
	}

	result, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidated{make(chan struct{}), int32(TX_VALIDATED_INIT), tx, 0, nil})
	if !loaded {
		result.tx = nil
		atomic.StoreInt32(&result.status, TX_VALIDATED_PROCCESSED)
		result.time = time.Now().Add(EXPIRE_TIME_MS).Unix()
		close(result.wait)
	}

	return nil
}

//blocking
func (validator *TxsValidator) ValidateTx(tx *transaction.Transaction) error {

	result, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidated{make(chan struct{}), TX_VALIDATED_INIT, tx, 0, nil})
	if !loaded {
		validator.processing.Store(tx.Bloom.HashStr, result)
	}

	<-result.wait

	return result.result
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
		&generics.Map[string, *txValidated]{},
		nil,
	}

	workers := make([]*TxsValidatorWorker, config.CPU_THREADS)
	for i := range workers {
		workers[i] = newTxsValidatorWorker(txsValidator.processing)
	}

	txsValidator.workers = workers

	for _, worker := range workers {
		worker.start()
	}

	go txsValidator.runRemoveExpiredTransactions()

	return txsValidator, nil
}
