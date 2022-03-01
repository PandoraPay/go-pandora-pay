package address_balance_decryptor

import (
	"pandora-pay/addresses"
	"pandora-pay/config"
	"pandora-pay/cryptography/crypto"
	"sync/atomic"
	"time"
)

type AddressBalanceDecryptorWorker struct {
	newWorkCn chan *addressBalanceDecryptorWork
}

func (worker *AddressBalanceDecryptorWorker) processWork(work *addressBalanceDecryptorWork) (uint64, error) {

	balancePoint, err := new(crypto.ElGamal).Deserialize(work.encryptedBalance)
	if err != nil {
		return 0, err
	}

	priv := &addresses.PrivateKey{work.privateKey}

	decrypted, err := priv.DecryptBalance(balancePoint, work.previousValue, work.ctx, work.statusCallback)
	if err != nil {
		return 0, err
	}

	return decrypted, nil
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
