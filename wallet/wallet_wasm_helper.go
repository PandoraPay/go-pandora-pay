package wallet

func (wallet *Wallet) GetPrivateKeys(publicKey, asset []byte) (privateKey []byte, previousValue uint64) {

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.addressesMap[string(publicKey)]

	if addr.PrivateKey != nil {
		privateKey = addr.PrivateKey.Key
	}

	return
}
