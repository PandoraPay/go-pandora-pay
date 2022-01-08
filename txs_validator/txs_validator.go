package txs_validator

import (
	"pandora-pay/blockchain/transactions/transaction"
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
	"time"
)

type txValidated struct {
	tx     *transaction.Transaction
	wait   chan struct{}
	time   *generics.Value[int64]
	result error
}

type TxsValidator struct {
	all        *generics.Map[string, *txValidated]
	processing *generics.Map[string, *txValidated]
	workers    []*TxsValidatorWorker
}

const (
	EXPIRE_TIME_MS = 10 * 60 * time.Second
)

func (validator *TxsValidator) MarkAsValidatedTx(tx *transaction.Transaction) error {

	if err := tx.BloomAll(); err != nil {
		return err
	}

	result, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidated{tx, make(chan struct{}), &generics.Value[int64]{}, nil})
	if !loaded {
		result.tx = nil
		result.time.Store(time.Now().Add(EXPIRE_TIME_MS).Unix())
		close(result.wait)
	}

	return nil
}

//blocking
func (validator *TxsValidator) ValidateTx(tx *transaction.Transaction) error {

	result, loaded := validator.all.LoadOrStore(tx.Bloom.HashStr, &txValidated{tx, make(chan struct{}), &generics.Value[int64]{}, nil})
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

			if t := work.time.Load(); t != 0 && t < now {
				validator.all.Delete(key)
			}

			c += 1
			if c%1000 == 0 {
				<-ticker.C
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
