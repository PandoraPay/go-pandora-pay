package address_balance_decryptor

import (
	"context"
	"github.com/tevino/abool"
	"pandora-pay/config"
	"pandora-pay/cryptography/bn256"
	"pandora-pay/cryptography/crypto"
	balance_decryptor "pandora-pay/cryptography/crypto/balance-decryptor"
	"pandora-pay/helpers/generics"
	"runtime"
)

type AddressBalanceDecryptor struct {
	all                   *generics.Map[string, *addressBalanceDecryptorWork]
	previousValues        *generics.Map[string, uint64]
	previousValuesChanged *abool.AtomicBool
	workers               []*AddressBalanceDecryptorWorker
	newWorkCn             chan *addressBalanceDecryptorWork
}

func (decryptor *AddressBalanceDecryptor) DecryptBalance(decryptionName string, publicKey, privateKey, encryptedBalance, asset []byte, useNewPreviousValue bool, newPreviousValue uint64, storeNewPreviousValue bool, ctx context.Context, statusCallback func(string)) (uint64, error) {

	if len(encryptedBalance) == 0 {
		return 0, nil
	}

	previousValue := uint64(0)
	if useNewPreviousValue {
		previousValue = newPreviousValue
	} else {
		previousValue, _ = decryptor.previousValues.Load(string(publicKey) + "_" + string(asset) + "_" + decryptionName)
	}

	balance, err := new(crypto.ElGamal).Deserialize(encryptedBalance)
	if err != nil {
		return 0, err
	}

	balancePoint := new(bn256.G1).Add(balance.Left, new(bn256.G1).Neg(new(bn256.G1).ScalarMult(balance.Right, new(crypto.BNRed).SetBytes(privateKey).BigInt())))
	if balance_decryptor.BalanceDecryptor.TryDecryptBalance(balancePoint, previousValue) {
		return previousValue, nil
	}

	foundWork, loaded := decryptor.all.LoadOrStore(string(publicKey)+"_"+string(encryptedBalance), &addressBalanceDecryptorWork{balancePoint, previousValue, make(chan struct{}), ADDRESS_BALANCE_DECRYPTED_INIT, 0, nil, ctx, statusCallback})
	if !loaded {
		decryptor.newWorkCn <- foundWork
	}

	<-foundWork.wait
	if foundWork.result.err != nil {
		return 0, foundWork.result.err
	}

	if storeNewPreviousValue {
		decryptor.previousValues.Store(string(publicKey)+"_"+string(asset)+"_"+decryptionName, foundWork.result.decryptedBalance)
		decryptor.previousValuesChanged.Set()
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
		abool.New(),
		make([]*AddressBalanceDecryptorWorker, threadsCount),
		make(chan *addressBalanceDecryptorWork, 1),
	}

	if runtime.GOARCH != "wasm" {
		if err := addressBalanceDecryptor.loadFromStore(); err != nil {
			return nil, err
		}
	}

	for i := range addressBalanceDecryptor.workers {
		addressBalanceDecryptor.workers[i] = newAddressBalanceDecryptorWorker(addressBalanceDecryptor.newWorkCn)
	}

	for _, worker := range addressBalanceDecryptor.workers {
		worker.start()
	}

	if runtime.GOARCH != "wasm" {
		go addressBalanceDecryptor.saveToStore()
	}

	return addressBalanceDecryptor, nil
}
