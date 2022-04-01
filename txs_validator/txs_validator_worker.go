package txs_validator

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/config"
	"sync/atomic"
	"time"
)

type TxsValidatorWorker struct {
	newValidationWorkCn chan *txValidatedWork
}

func (worker *TxsValidatorWorker) verifyTx(foundWork *txValidatedWork) error {

	if err := foundWork.tx.VerifyBloomAll(); err != nil {
		return err
	}

	hashForSignature := foundWork.tx.GetHashSigningManually()

	switch foundWork.tx.Version {
	case transaction_type.TX_SIMPLE:
		base := foundWork.tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
		if !base.VerifySignatureManually(hashForSignature) {
			return errors.New("Signature Verified failed")
		}
	default:
		return errors.New("Invalid Tx Version")
	}

	return nil
}

func (worker *TxsValidatorWorker) run() {

	for {
		foundWork, _ := <-worker.newValidationWorkCn

		if err := foundWork.tx.BloomAll(); err != nil {
			foundWork.result = err
		} else {
			foundWork.bloomExtra = foundWork.tx.TransactionBaseInterface.GetBloomExtra()
			if err = worker.verifyTx(foundWork); err != nil {
				foundWork.result = err
			}
		}

		foundWork.tx = nil
		foundWork.time = time.Now().Add(EXPIRE_TIME_MS).Unix()
		atomic.StoreInt32(&foundWork.status, TX_VALIDATED_PROCCESSED)

		close(foundWork.wait)

		if config.LIGHT_COMPUTATIONS {
			time.Sleep(50 * time.Millisecond)
		}

	}
}

func (worker *TxsValidatorWorker) start() {
	go worker.run()
}

func newTxsValidatorWorker(newValidationWorkCn chan *txValidatedWork) *TxsValidatorWorker {
	worker := &TxsValidatorWorker{
		newValidationWorkCn,
	}
	return worker
}
