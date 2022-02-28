package address_balance_decryptor

import (
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
)

type AddressBalanceDecryptor struct {
	all                 *generics.Map[string, *addressBalanceDecryptedWork]
	workers             []*AddressBalanceDecryptorWorker
	newValidationWorkCn chan *addressBalanceDecryptedWork
}

func (decryptor *AddressBalanceDecryptor) DecryptBalanceByPrivateKey(privateKey []byte) error {
	return nil
}

func (decryptor *AddressBalanceDecryptor) DecryptBalance(publicKey, privateKey []byte, balance []byte) (uint64, error) {

	foundWork, loaded := decryptor.all.LoadOrStore(string(publicKey)+"_"+string(balance), &addressBalanceDecryptedWork{make(chan struct{}), ADDRESS_BALANCE_DECRYPTED_INIT, 0, 0, nil})
	if !loaded {
		decryptor.newValidationWorkCn <- foundWork
	}

	<-foundWork.wait
	if foundWork.result != nil {
		return 0, foundWork.result
	}

	return foundWork.decrypted, nil
}

func NewAddressBalanceDecryptor() (*AddressBalanceDecryptor, error) {

	threadsCount := config.CPU_THREADS
	if config.LIGHT_COMPUTATIONS {
		threadsCount = generics.Max(1, config.CPU_THREADS/2)
	}

	addressBalanceDecryptor := &AddressBalanceDecryptor{
		&generics.Map[string, *addressBalanceDecryptedWork]{},
		make([]*AddressBalanceDecryptorWorker, threadsCount),
		make(chan *addressBalanceDecryptedWork, 1),
	}

	for i := range addressBalanceDecryptor.workers {
		addressBalanceDecryptor.workers[i] = newAddressBalanceDecryptorWorker(addressBalanceDecryptor.newValidationWorkCn)
	}

	for _, worker := range addressBalanceDecryptor.workers {
		worker.start()
	}

	return addressBalanceDecryptor, nil
}
