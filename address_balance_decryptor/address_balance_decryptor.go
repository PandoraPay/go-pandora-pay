package address_balance_decryptor

import (
	"pandora-pay/addresses"
	"pandora-pay/config"
	"pandora-pay/helpers/generics"
)

type AddressBalanceDecryptor struct {
	all            *generics.Map[string, *addressBalanceDecryptorWork]
	previousValues *generics.Map[string, uint64]
	workers        []*AddressBalanceDecryptorWorker
	newWorkCn      chan *addressBalanceDecryptorWork
}

func (decryptor *AddressBalanceDecryptor) DecryptBalanceByPrivateKey(privateKey, encryptedBalance []byte, usePreviousValue, storeNewPreviousValue bool) (uint64, error) {

	priv := &addresses.PrivateKey{privateKey}

	return decryptor.DecryptBalance(priv.GeneratePublicKey(), privateKey, encryptedBalance, usePreviousValue, storeNewPreviousValue)
}

func (decryptor *AddressBalanceDecryptor) DecryptBalance(publicKey, privateKey []byte, encryptedBalance []byte, usePreviousValue, storeNewPreviousValue bool) (uint64, error) {

	if len(encryptedBalance) == 0 {
		return 0, nil
	}

	previousValue := uint64(0)
	if usePreviousValue {
		previousValue, _ = decryptor.previousValues.Load(string(publicKey))
	}

	foundWork, loaded := decryptor.all.LoadOrStore(string(publicKey)+"_"+string(encryptedBalance), &addressBalanceDecryptorWork{encryptedBalance, privateKey, previousValue, make(chan struct{}), ADDRESS_BALANCE_DECRYPTED_INIT, 0, nil})
	if !loaded {
		decryptor.newWorkCn <- foundWork
	}

	<-foundWork.wait
	if foundWork.result.err != nil {
		return 0, foundWork.result.err
	}

	if storeNewPreviousValue {
		decryptor.previousValues.Store(string(publicKey), foundWork.result.decryptedBalance)
	}

	return foundWork.result.decryptedBalance, nil
}

func NewAddressBalanceDecryptor() (*AddressBalanceDecryptor, error) {

	threadsCount := config.CPU_THREADS
	if config.LIGHT_COMPUTATIONS {
		threadsCount = generics.Max(1, config.CPU_THREADS/2)
	}

	addressBalanceDecryptor := &AddressBalanceDecryptor{
		&generics.Map[string, *addressBalanceDecryptorWork]{},
		&generics.Map[string, uint64]{},
		make([]*AddressBalanceDecryptorWorker, threadsCount),
		make(chan *addressBalanceDecryptorWork, 1),
	}

	for i := range addressBalanceDecryptor.workers {
		addressBalanceDecryptor.workers[i] = newAddressBalanceDecryptorWorker(addressBalanceDecryptor.newWorkCn)
	}

	for _, worker := range addressBalanceDecryptor.workers {
		worker.start()
	}

	return addressBalanceDecryptor, nil
}
