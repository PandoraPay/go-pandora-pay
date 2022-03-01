package wallet

func (wallet *Wallet) GetDataForDecryptingBalance(publicKey, asset []byte) (privateKey []byte, previousValue uint64) {

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.addressesMap[string(publicKey)]
	privateKey = addr.PrivateKey.Key

	return
}
