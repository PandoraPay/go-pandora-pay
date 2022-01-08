package txs_validator

import (
	"errors"
	"pandora-pay/blockchain/transactions/transaction/transaction_simple"
	"pandora-pay/blockchain/transactions/transaction/transaction_type"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether"
	"pandora-pay/blockchain/transactions/transaction/transaction_zether/transaction_zether_payload"
	"pandora-pay/helpers/generics"
	"time"
)

type TxsValidatorWorker struct {
	processing *generics.Map[string, *txValidated]
}

func (worker *TxsValidatorWorker) verifyTx(foundWork *txValidated) error {

	if err := foundWork.tx.BloomAll(); err != nil {
		return err
	}

	if err := foundWork.tx.VerifyBloomAll(); err != nil {
		return err
	}

	hashForSignature := foundWork.tx.GetHashSigningManually()

	switch foundWork.tx.Version {
	case transaction_type.TX_SIMPLE:
		base := foundWork.tx.TransactionBaseInterface.(*transaction_simple.TransactionSimple)
		if !base.VerifySignatureManually(hashForSignature) {

		}

	case transaction_type.TX_ZETHER:
		base := foundWork.tx.TransactionBaseInterface.(*transaction_zether.TransactionZether)
		//verify signature
		assetMap := map[string]int{}
		for _, payload := range base.Payloads {
			if payload.Proof.Verify(payload.Asset, assetMap[string(payload.Asset)], base.ChainHash, payload.Statement, hashForSignature, payload.BurnValue) == false {
				return errors.New("Zether Failed for Transaction")
			}
			assetMap[string(payload.Asset)] = assetMap[string(payload.Asset)] + 1
		}

		for _, payload := range base.Payloads {
			switch payload.PayloadScript {
			case transaction_zether_payload.SCRIPT_DELEGATE_STAKE, transaction_zether_payload.SCRIPT_CLAIM,
				transaction_zether_payload.SCRIPT_ASSET_SUPPLY_INCREASE:
				if payload.Extra.VerifyExtraSignature(hashForSignature) == false {
					return errors.New("DelegatedPublicKey signature failed")
				}
			}
		}
	default:
		return errors.New("Invalid Tx Version")
	}

	return nil
}

func (worker *TxsValidatorWorker) run() {

	ticker := time.NewTicker(10 * time.Millisecond)

	for {
		var foundWork *txValidated
		worker.processing.Range(func(key string, value *txValidated) bool {
			foundWork = value
			return false
		})

		if foundWork == nil {
			<-ticker.C
			continue
		}

		if err := worker.verifyTx(foundWork); err != nil {
			foundWork.result = err
		} else {

		}

		foundWork.tx = nil
		foundWork.time.Store(time.Now().Add(EXPIRE_TIME_MS).Unix())

		worker.processing.Delete(foundWork.tx.Bloom.HashStr)

		close(foundWork.wait)

	}
}

func (worker *TxsValidatorWorker) start() {
	go worker.run()
}

func newTxsValidatorWorker(processing *generics.Map[string, *txValidated]) *TxsValidatorWorker {
	worker := &TxsValidatorWorker{
		processing,
	}
	return worker
}
