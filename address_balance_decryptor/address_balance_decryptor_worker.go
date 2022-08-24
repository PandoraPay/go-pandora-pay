package address_balance_decryptor

import (
	"pandora-pay/config"
	"pandora-pay/cryptography/crypto/balance_decryptor"
	"sync/atomic"
	"time"
)

type AddressBalanceDecryptorWorker struct {
	newWorkCn chan *addressBalanceDecryptorWork
}

func (worker *AddressBalanceDecryptorWorker) processWork(work *addressBalanceDecryptorWork) (uint64, error) {
	return balance_decryptor.BalanceDecryptor.DecryptBalance(work.encryptedBalance, false, 0, work.ctx, work.statusCallback)
}

func (worker *AddressBalanceDecryptorWorker) run() {

	for {
		foundWork, _ := <-worker.newWorkCn

		foundWork.result = &addressBalanceDecryptorWorkResult{}

		foundWork.result.decryptedBalance, foundWork.result.err = worker.processWork(foundWork)

		foundWork.time = time.Now().Unix()
		atomic.StoreInt32(&foundWork.status, ADDRESS_BALANCE_DECRYPTED_PROCESSED)

		close(foundWork.wait)

		if config.LIGHT_COMPUTATIONS {
			time.Sleep(50 * time.Millisecond)
		}

	}
}

func (worker *AddressBalanceDecryptorWorker) start() {
	go worker.run()
}

func newAddressBalanceDecryptorWorker(newWorkCn chan *addressBalanceDecryptorWork) *AddressBalanceDecryptorWorker {
	worker := &AddressBalanceDecryptorWorker{
		newWorkCn,
	}
	return worker
}
