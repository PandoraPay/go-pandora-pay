package address_balance_decryptor

import (
	"pandora-pay/config"
	"sync/atomic"
	"time"
)

type AddressBalanceDecryptorWorker struct {
	newAddressBalanceDecryptedWorkCn chan *addressBalanceDecryptedWork
}

func (worker *AddressBalanceDecryptorWorker) run() {

	for {
		foundWork, _ := <-worker.newAddressBalanceDecryptedWorkCn

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

func newAddressBalanceDecryptorWorker(newWorkCn chan *addressBalanceDecryptedWork) *AddressBalanceDecryptorWorker {
	worker := &AddressBalanceDecryptorWorker{
		newWorkCn,
	}
	return worker
}
