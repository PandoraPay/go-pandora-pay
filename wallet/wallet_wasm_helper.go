package wallet

func (wallet *Wallet) GetPrivateKeys(publicKey, asset []byte) (privateKey, spendPrivateKey []byte, previousValue uint64) {

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.addressesMap[string(publicKey)]

	if addr.PrivateKey != nil {
		privateKey = addr.PrivateKey.Key
	}

	if addr.SpendPrivateKey != nil {
		spendPrivateKey = addr.SpendPrivateKey.Key
	}

	return
}
