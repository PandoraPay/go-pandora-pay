package wallet

func (wallet *Wallet) GetPrivateKeys(publicKeyHash, asset []byte) (privateKey []byte) {

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.addressesMap[string(publicKeyHash)]

	if addr.PrivateKey != nil {
		privateKey = addr.PrivateKey.Key
	}

	return
}
