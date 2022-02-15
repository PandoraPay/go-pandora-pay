package wallet

import (
	"encoding/base64"
)

func (wallet *Wallet) GetDataForDecryptingBalance(publicKey, asset []byte) (privateKey []byte, previousValue uint64) {

	wallet.Lock.RLock()
	defer wallet.Lock.RUnlock()

	addr := wallet.addressesMap[string(publicKey)]
	privateKey = addr.PrivateKey.Key

	if addr.DecryptedBalances[base64.StdEncoding.EncodeToString(asset)] != nil {
		previousValue = addr.DecryptedBalances[base64.StdEncoding.EncodeToString(asset)].Amount
	}

	return
}
